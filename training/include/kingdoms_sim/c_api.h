#ifndef KINGDOMS_SIM_C_API_H
#define KINGDOMS_SIM_C_API_H

// Pure-C boundary for RL training backends (PufferLib's binding.c is compiled
// as C). Everything tensor-facing is flat float arrays; the C++ Observation/
// Action structs never cross this line. See c_api.cpp for the layout.

#ifdef __cplusplus
extern "C" {
#endif

// One-hot table sizes. Runtime-checked against the loaded assets in
// kd_create (abort with a message rather than silently mis-indexing if
// items.json/npcs.json grow past these).
#define KD_NUM_ITEMS 28
#define KD_NUM_NPC_TYPES 19

// Flattened observation layout (all float32, in this order):
//   self:       2 pos + 1 hp-frac + 1 max-hp + 1 cooldown + 4 stats      = 9
//   equipment:  hand/head/body one-hot, KD_NUM_ITEMS each               = 84
//   inventory:  24 slots x KD_NUM_ITEMS one-hot (empty = all zero)     = 672
//   npcs:       16 x (present, rel x, rel y, hp) + KD_NUM_NPC_TYPES    = 368
//   projectiles:24 x (present, rel x, rel y, damage, is_enemy)         = 120
//   bombs:      12 x (present, rel x, rel y, damage, radius, timer)     = 72
//   loot:       12 x (present, rel x, rel y) + KD_NUM_ITEMS one-hot    = 372
#define KD_OBS_SIZE 1697

// Discrete action heads (indices arrive as floats from the trainer):
//   [0] move:  0 = stay, 1..8 = compass direction (i-1) * 45 degrees
//   [1] aim:   0 = no attack, 1..16 = attack toward (i-1) * 22.5 degrees
//   [2] use:   0 = no-op, 1..24 = act on backpack slot i-1 -- head/body
//              items equip to their slot, hand items/consumables become the
//              active hand slot. Empty slots are no-ops.
#define KD_NUM_ACTION_HEADS 3
#define KD_ACT_SIZES {9, 17, 25}

// Mirrors kingdoms::RewardWeights (reward.hpp) across the C boundary so
// training configs can tune rewards without recompiling this library.
typedef struct {
  float damage_dealt;
  float damage_taken;
  float kill;
  float loot_pickup;
  float gear_upgrade;
  float attack_miss;
  float death;
} KdRewardWeights;

typedef struct {
  float episode_return;
  float episode_length;
  float kills;
  float damage_dealt;
  float damage_taken;
  float loot_picked_up;
  float gear_score;  // episode high-water mark (Environment::FinalGearScore)
  float died;        // 1 = death, 0 = max_ticks timeout
} KdEpisodeStats;

// Compact draw-ready snapshot of the last observation, for a shapes-only
// eval visualizer. Positions are absolute world coordinates.
typedef struct {
  float x, y;
  int type;      // 1 npc, 2 enemy projectile, 3 player projectile, 4 bomb, 5 loot
  int id;        // npc type id (npcs) / item id (loot), else 0
  float health;  // npcs only
  float radius;  // bombs: blast radius
} KdRenderEntity;

typedef struct {
  float player_x, player_y;
  float health, max_health;
  float gear_score;
  int num_entities;
  KdRenderEntity entities[64];  // kMaxObservedEntities
} KdRenderState;

void kd_render_state(void* handle, KdRenderState* out);

// Loads the shared AssetDatabase on first call (thread-safe).
// `weights` may be NULL for the reward.hpp defaults.
void* kd_create(unsigned int seed, int max_ticks, const KdRewardWeights* weights);
void kd_destroy(void* handle);
void kd_reset(void* handle, unsigned int seed, float* obs_out);

// Advances one tick. `actions` is KD_NUM_ACTION_HEADS floats (head indices).
// Writes KD_OBS_SIZE floats and the tick's reward. Returns 1 when the episode
// ended this tick (death or timeout) -- `stats_out` is only filled then, and
// the caller is expected to kd_reset before the next kd_step.
int kd_step(void* handle, const float* actions, float* obs_out, float* reward_out,
            KdEpisodeStats* stats_out);

#ifdef __cplusplus
}
#endif

#endif  // KINGDOMS_SIM_C_API_H
