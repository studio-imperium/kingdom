#include "kingdoms_sim/c_api.h"

#include <cmath>
#include <cstdio>
#include <cstdlib>
#include <cstring>

#include "kingdoms_sim/environment.hpp"

// See c_api.h for the observation/action layout this file implements.

namespace kingdoms {
namespace {

constexpr float kArenaScale = 200.0f;  // player abs position / arena radius
constexpr float kRelScale = 100.0f;    // entity positions relative to player
constexpr float kHealthScale = 100.0f;
constexpr float kDamageScale = 50.0f;
constexpr float kRadiusScale = 10.0f;
constexpr float kTimerScale = 3.0f;
constexpr float kMaxHealthScale = 200.0f;

const AssetDatabase& SharedAssets() {
  // Function-local static: loaded once, thread-safe under C++11 init rules.
  static const AssetDatabase db = AssetDatabase::LoadFromDirectory(KINGDOMS_ASSETS_DIR);
  return db;
}

inline void OneHotItem(float* out, int item_id) {
  // Caller zeroed the buffer; item_id -1 (empty) leaves all zeros.
  if (item_id >= 0 && item_id < KD_NUM_ITEMS) out[item_id] = 1.0f;
}

int Flatten(const Observation& obs, float* out) {
  std::memset(out, 0, KD_OBS_SIZE * sizeof(float));
  int i = 0;

  out[i++] = obs.player_pos.x / kArenaScale;
  out[i++] = obs.player_pos.y / kArenaScale;
  out[i++] = obs.max_health > 0.0f ? obs.health / obs.max_health : 0.0f;
  out[i++] = obs.max_health / kMaxHealthScale;
  out[i++] = obs.attack_cooldown;
  out[i++] = obs.speed;
  out[i++] = obs.power;
  out[i++] = obs.reload;
  out[i++] = obs.regen;

  OneHotItem(&out[i], obs.hand); i += KD_NUM_ITEMS;
  OneHotItem(&out[i], obs.head); i += KD_NUM_ITEMS;
  OneHotItem(&out[i], obs.body); i += KD_NUM_ITEMS;

  for (int s = 0; s < kBackpackSlots; s++) {
    OneHotItem(&out[i], obs.inventory[s]);
    i += KD_NUM_ITEMS;
  }

  // Entity slots are typed by fixed ranges (see observation.hpp), so the
  // per-type feature widths below can differ without ambiguity.
  for (int s = 0; s < kMaxNpcSlots; s++) {
    const ObservedEntity& e = obs.entities[s];
    if (e.type == ObservedEntityType::Npc) {
      out[i] = 1.0f;
      out[i + 1] = (e.pos.x - obs.player_pos.x) / kRelScale;
      out[i + 2] = (e.pos.y - obs.player_pos.y) / kRelScale;
      out[i + 3] = e.npc_health / kHealthScale;
      if (e.npc_type_id < KD_NUM_NPC_TYPES) out[i + 4 + e.npc_type_id] = 1.0f;
    }
    i += 4 + KD_NUM_NPC_TYPES;
  }

  for (int s = 0; s < kMaxProjectileSlots; s++) {
    const ObservedEntity& e = obs.entities[kMaxNpcSlots + s];
    if (e.type == ObservedEntityType::Projectile) {
      out[i] = 1.0f;
      out[i + 1] = (e.pos.x - obs.player_pos.x) / kRelScale;
      out[i + 2] = (e.pos.y - obs.player_pos.y) / kRelScale;
      out[i + 3] = e.projectile_damage / kDamageScale;
      out[i + 4] = e.is_enemy_projectile ? 1.0f : 0.0f;
    }
    i += 5;
  }

  for (int s = 0; s < kMaxBombSlots; s++) {
    const ObservedEntity& e = obs.entities[kMaxNpcSlots + kMaxProjectileSlots + s];
    if (e.type == ObservedEntityType::Bomb) {
      out[i] = 1.0f;
      out[i + 1] = (e.pos.x - obs.player_pos.x) / kRelScale;
      out[i + 2] = (e.pos.y - obs.player_pos.y) / kRelScale;
      out[i + 3] = e.bomb_damage / kDamageScale;
      out[i + 4] = e.bomb_radius / kRadiusScale;
      out[i + 5] = e.bomb_timer / kTimerScale;
    }
    i += 6;
  }

  for (int s = 0; s < kMaxLootSlots; s++) {
    const ObservedEntity& e =
        obs.entities[kMaxNpcSlots + kMaxProjectileSlots + kMaxBombSlots + s];
    if (e.type == ObservedEntityType::Loot) {
      out[i] = 1.0f;
      out[i + 1] = (e.pos.x - obs.player_pos.x) / kRelScale;
      out[i + 2] = (e.pos.y - obs.player_pos.y) / kRelScale;
      OneHotItem(&out[i + 3], e.loot_item_id);
    }
    i += 3 + KD_NUM_ITEMS;
  }

  return i;
}

struct CApiEnv {
  Environment env;
  Observation last_obs;  // action translation reads this (what the policy saw)
  KdEpisodeStats acc{};

  CApiEnv(unsigned int seed, int max_ticks, RewardWeights weights)
      : env(SharedAssets(), seed, weights, max_ticks) {}
};

Action Translate(const CApiEnv& e, const float* heads) {
  Action a;
  int move = (int)heads[0];
  int aim = (int)heads[1];
  int use = (int)heads[2];

  if (move >= 1 && move <= 8) {
    float ang = (float)(move - 1) * (kPi / 4.0f);
    a.move = {std::cos(ang), std::sin(ang)};
  }
  if (aim >= 1 && aim <= 16) {
    float ang = (float)(aim - 1) * (kPi / 8.0f);
    a.attack = true;
    a.attack_direction = {std::cos(ang), std::sin(ang)};
  }
  if (use >= 1 && use <= kBackpackSlots) {
    uint8_t slot = (uint8_t)(use - 1);
    int item_id = e.last_obs.inventory[slot];
    if (item_id >= 0) {
      ItemSlot kind = SharedAssets().Item((uint8_t)item_id).slot;
      if (kind == ItemSlot::Head) {
        a.do_equip = true; a.equip_from = slot; a.equip_to = kHeadSlot;
      } else if (kind == ItemSlot::Body) {
        a.do_equip = true; a.equip_from = slot; a.equip_to = kBodySlot;
      } else {
        a.do_switch = true; a.switch_to_slot = slot;
      }
    }
  }
  return a;
}

}  // namespace
}  // namespace kingdoms

extern "C" {

void* kd_create(unsigned int seed, int max_ticks, const KdRewardWeights* weights) {
  using namespace kingdoms;
  RewardWeights rw;  // reward.hpp defaults
  if (weights != NULL) {
    rw.damage_dealt = weights->damage_dealt;
    rw.damage_taken = weights->damage_taken;
    rw.kill = weights->kill;
    rw.loot_pickup = weights->loot_pickup;
    rw.gear_upgrade = weights->gear_upgrade;
    rw.attack_miss = weights->attack_miss;
    rw.death = weights->death;
  }
  const AssetDatabase& db = SharedAssets();
  for (const NpcData& npc : db.AllNpcs()) {
    if (npc.id >= KD_NUM_NPC_TYPES) {
      std::fprintf(stderr, "kd_create: npc id %d >= KD_NUM_NPC_TYPES (%d); "
                   "bump c_api.h and rebuild\n", npc.id, KD_NUM_NPC_TYPES);
      std::abort();
    }
  }
  auto* e = new CApiEnv(seed, max_ticks, rw);
  float* scratch = (float*)std::calloc(KD_OBS_SIZE, sizeof(float));
  int n = Flatten(e->last_obs, scratch);
  std::free(scratch);
  if (n != KD_OBS_SIZE) {
    std::fprintf(stderr, "kd_create: flattened size %d != KD_OBS_SIZE %d\n", n, KD_OBS_SIZE);
    std::abort();
  }
  return e;
}

void kd_destroy(void* handle) { delete (kingdoms::CApiEnv*)handle; }

void kd_render_state(void* handle, KdRenderState* out) {
  using namespace kingdoms;
  static_assert(kMaxObservedEntities == 64, "bump KdRenderState.entities and this assert");
  auto* e = (CApiEnv*)handle;
  const Observation& o = e->last_obs;
  out->player_x = o.player_pos.x;
  out->player_y = o.player_pos.y;
  out->health = o.health;
  out->max_health = o.max_health;
  out->gear_score = e->env.FinalGearScore();
  int n = 0;
  for (int s = 0; s < kMaxObservedEntities; s++) {
    const ObservedEntity& ent = o.entities[s];
    if (ent.type == ObservedEntityType::None) continue;
    KdRenderEntity* r = &out->entities[n++];
    r->x = ent.pos.x;
    r->y = ent.pos.y;
    r->id = 0;
    r->health = 0.0f;
    r->radius = 0.0f;
    switch (ent.type) {
      case ObservedEntityType::Npc:
        r->type = 1; r->id = ent.npc_type_id; r->health = ent.npc_health; break;
      case ObservedEntityType::Projectile:
        r->type = ent.is_enemy_projectile ? 2 : 3; break;
      case ObservedEntityType::Bomb:
        r->type = 4; r->radius = ent.bomb_radius; break;
      case ObservedEntityType::Loot:
        r->type = 5; r->id = ent.loot_item_id; break;
      default:
        n--; break;
    }
  }
  out->num_entities = n;
}

void kd_reset(void* handle, unsigned int seed, float* obs_out) {
  auto* e = (kingdoms::CApiEnv*)handle;
  e->last_obs = e->env.Reset(seed);
  e->acc = KdEpisodeStats{};
  kingdoms::Flatten(e->last_obs, obs_out);
}

int kd_step(void* handle, const float* actions, float* obs_out, float* reward_out,
            KdEpisodeStats* stats_out) {
  using namespace kingdoms;
  auto* e = (CApiEnv*)handle;

  StepResult r = e->env.Step(Translate(*e, actions));
  e->last_obs = r.observation;
  Flatten(r.observation, obs_out);
  *reward_out = r.reward;

  e->acc.episode_return += r.reward;
  e->acc.episode_length += 1.0f;
  e->acc.kills += (float)r.events.kills;
  e->acc.damage_dealt += r.events.damage_dealt;
  e->acc.damage_taken += r.events.damage_taken;
  e->acc.loot_picked_up += (float)r.events.loot_picked_up.size();

  if (r.done) {
    e->acc.gear_score = e->env.FinalGearScore();
    e->acc.died = r.events.player_died ? 1.0f : 0.0f;
    *stats_out = e->acc;
    return 1;
  }
  return 0;
}

}  // extern "C"
