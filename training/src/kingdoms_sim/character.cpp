#include "kingdoms_sim/character.hpp"

namespace kingdoms {

void Character::Tick(float dt_seconds) {
  attack_cooldown -= dt_seconds;
  if (attack_cooldown < -1.0f) attack_cooldown = -1.0f;

  health += regen * dt_seconds;
  if (health > max_health) health = max_health;
}

void Character::Damage(float amount) { health -= amount; }

void Character::ApplyGear(const AssetDatabase& assets) {
  const ItemData& helmet = assets.Item(head);
  const ItemData& torso = assets.Item(body);

  max_health = 25.0f * helmet.stats.health * torso.stats.health;
  regen = helmet.stats.regen * torso.stats.regen;
  speed = helmet.stats.speed * torso.stats.speed;
  power = helmet.stats.damage * torso.stats.damage;
  reload = helmet.stats.reload * torso.stats.reload;
}

bool Character::AddItemOrBust(uint8_t item_id) {
  for (uint8_t i = 0; i < kBackpackSlots; ++i) {
    if (inventory.find(i) == inventory.end()) {
      inventory[i] = item_id;
      return true;
    }
  }
  return false;
}

void Character::ChangeInventory(const AssetDatabase& assets, uint8_t to, uint8_t from) {
  if (to == from) return;

  auto is_gear = [](uint8_t slot) { return slot == kHeadSlot || slot == kBodySlot; };

  if (is_gear(to)) {
    uint8_t& to_gear = (to == kHeadSlot) ? head : body;
    ItemSlot to_type = (to == kHeadSlot) ? ItemSlot::Head : ItemSlot::Body;
    uint8_t to_default = (to == kHeadSlot) ? kDefaultHeadItem : kDefaultBodyItem;

    auto from_it = inventory.find(from);
    if (from_it == inventory.end() || assets.Item(from_it->second).slot != to_type) return;

    uint8_t from_item = from_it->second;
    uint8_t equipped = to_gear;
    to_gear = from_item;

    if (equipped == to_default) {
      inventory.erase(from);
    } else {
      inventory[from] = equipped;
    }
    ApplyGear(assets);
    return;
  }

  if (is_gear(from)) {
    uint8_t& from_gear = (from == kHeadSlot) ? head : body;
    ItemSlot from_type = (from == kHeadSlot) ? ItemSlot::Head : ItemSlot::Body;
    uint8_t from_default = (from == kHeadSlot) ? kDefaultHeadItem : kDefaultBodyItem;

    uint8_t equipped = from_gear;
    if (equipped == from_default) return;

    auto to_it = inventory.find(to);
    if (to_it != inventory.end()) {
      if (assets.Item(to_it->second).slot != from_type) return;
      from_gear = to_it->second;
    } else {
      from_gear = from_default;
    }
    inventory[to] = equipped;
    ApplyGear(assets);
    return;
  }

  auto from_it = inventory.find(from);
  if (from_it == inventory.end()) return;
  uint8_t item1 = from_it->second;

  auto to_it = inventory.find(to);
  if (to_it != inventory.end()) {
    inventory[from] = to_it->second;
  } else {
    inventory.erase(from);
  }
  inventory[to] = item1;
  ApplyGear(assets);
}

void Character::RemoveInventory(const AssetDatabase& assets, uint8_t slot) {
  if (slot < kBackpackSlots) {
    inventory.erase(slot);
  } else if (slot == kHeadSlot) {
    head = kDefaultHeadItem;
  } else {
    body = kDefaultBodyItem;
  }
  ApplyGear(assets);
}

Character MakeDefaultCharacter(const AssetDatabase& assets) {
  Character c;
  c.hand = 0;
  c.head = kDefaultHeadItem;
  c.body = kDefaultBodyItem;
  c.inventory[0] = 8;  // Kendo Stick; matches DefaultCharacter's starting weapon in character.go
  c.ApplyGear(assets);
  c.health = c.max_health;
  return c;
}

}  // namespace kingdoms
