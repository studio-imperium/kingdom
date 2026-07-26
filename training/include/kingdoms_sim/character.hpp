#pragma once

#include <cstdint>
#include <unordered_map>

#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/types.hpp"

// Mirrors server/engine/character.go, minus everything that only existed to
// serialize state over the network (send channel, PackFull/Pack, angle --
// angle was a cosmetic facing byte set directly from client input and is
// dropped entirely; see world.hpp for how attacks pick a direction instead).

namespace kingdoms {

struct Character {
  Vec2 pos;

  uint8_t hand = 0;
  uint8_t head = kDefaultHeadItem;
  uint8_t body = kDefaultBodyItem;

  float max_health = 25.0f;
  float health = 25.0f;
  float regen = 1.0f;
  float speed = 1.0f;
  float power = 1.0f;
  float reload = 1.0f;

  std::unordered_map<uint8_t, uint8_t> inventory;  // backpack slots [0, kBackpackSlots)
  uint8_t attack_counter = 0;
  float attack_cooldown = 0.0f;
  bool dead = false;

  // Cooldown decay + passive regen. Mirrors Character.Tick in character.go.
  void Tick(float dt_seconds);

  // Subtracts health. Does not check/mark death -- World does that once per
  // tick so it can attribute the kill/episode-end consistently.
  void Damage(float amount);

  // Recomputes max_health/regen/speed/power/reload from equipped head/body
  // gear. Mirrors Character.Apply in character.go.
  void ApplyGear(const AssetDatabase& assets);

  // Fills the first empty backpack slot with `item_id`; returns false if full.
  // Mirrors AddItemOrBust in loot.go.
  bool AddItemOrBust(uint8_t item_id);

  // Swap/equip between two slots (backpack indices, or kHeadSlot/kBodySlot).
  // No-ops on invalid moves exactly like ChangeInventory in character.go
  // (e.g. equipping a body item into the head slot).
  void ChangeInventory(const AssetDatabase& assets, uint8_t to, uint8_t from);

  // Clears a backpack slot, or resets head/body to its default item. Mirrors
  // RemoveInventory in character.go.
  void RemoveInventory(const AssetDatabase& assets, uint8_t slot);
};

Character MakeDefaultCharacter(const AssetDatabase& assets);

}  // namespace kingdoms
