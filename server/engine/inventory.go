package engine

import "kingdoms/engine/assets"

func (c *character) selectSlot(slot uint8) bool {
	if slot >= inventorySlots {
		return false
	}
	c.hand = slot
	return true
}

func (c *character) changeInventory(catalog *assets.Catalog, to, from uint8) bool {
	if to == from || to > bodySlot || from > bodySlot {
		return false
	}

	if to == headSlot || to == bodySlot {
		return c.equip(catalog, to, from)
	}
	if from == headSlot || from == bodySlot {
		return c.unequip(catalog, to, from)
	}

	item, exists := c.inventory[from]
	if !exists {
		return false
	}

	if other, occupied := c.inventory[to]; occupied {
		c.inventory[from] = other
	} else {
		delete(c.inventory, from)
	}
	c.inventory[to] = item
	c.apply(catalog)
	return true
}

func (c *character) equip(catalog *assets.Catalog, to, from uint8) bool {
	itemID, exists := c.inventory[from]
	if !exists {
		return false
	}
	item, valid := catalog.Item(itemID)
	if !valid || item.Slot != equipmentType(to) {
		return false
	}

	equipped := c.equipped(to)
	c.setEquipped(to, itemID)
	if equipped == equipmentDefault(to) {
		delete(c.inventory, from)
	} else {
		c.inventory[from] = equipped
	}
	c.apply(catalog)
	return true
}

func (c *character) unequip(catalog *assets.Catalog, to, from uint8) bool {
	equipped := c.equipped(from)
	if equipped == equipmentDefault(from) {
		return false
	}

	if itemID, occupied := c.inventory[to]; occupied {
		item, valid := catalog.Item(itemID)
		if !valid || item.Slot != equipmentType(from) {
			return false
		}
		c.setEquipped(from, itemID)
	} else {
		c.setEquipped(from, equipmentDefault(from))
	}

	c.inventory[to] = equipped
	c.apply(catalog)
	return true
}

func (c *character) drop(catalog *assets.Catalog, slot uint8) (uint8, bool) {
	var item uint8

	switch slot {
	case headSlot, bodySlot:
		item = c.equipped(slot)
		if item == equipmentDefault(slot) {
			return 0, false
		}
		c.setEquipped(slot, equipmentDefault(slot))
	default:
		if slot >= inventorySlots {
			return 0, false
		}
		var exists bool
		item, exists = c.inventory[slot]
		if !exists {
			return 0, false
		}
		delete(c.inventory, slot)
	}

	c.apply(catalog)
	return item, true
}

func (c *character) consume(catalog *assets.Catalog, slot uint8) bool {
	if slot >= inventorySlots {
		return false
	}
	if _, exists := c.inventory[slot]; !exists {
		return false
	}
	delete(c.inventory, slot)
	c.apply(catalog)
	return true
}

func (c *character) equipped(slot uint8) uint8 {
	if slot == headSlot {
		return c.head
	}
	return c.body
}

func (c *character) setEquipped(slot, item uint8) {
	if slot == headSlot {
		c.head = item
	} else {
		c.body = item
	}
}

func equipmentType(slot uint8) string {
	if slot == headSlot {
		return "head"
	}
	return "body"
}

func equipmentDefault(slot uint8) uint8 {
	if slot == headSlot {
		return defaultHead
	}
	return defaultBody
}
