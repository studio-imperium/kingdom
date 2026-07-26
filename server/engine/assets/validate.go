package assets

import "fmt"

func (c *Catalog) validate() error {
	for id, npc := range c.npcs {
		if int(npc.ID) != id {
			return fmt.Errorf("npc %d has id %d", id, npc.ID)
		}
		if len(npc.Modes) == 0 {
			return fmt.Errorf("npc %d has no modes", id)
		}
		if int(npc.Loot) >= len(c.loot) {
			return fmt.Errorf("npc %d has invalid loot table %d", id, npc.Loot)
		}
		for _, mode := range npc.Modes {
			if !validMovement(mode.Movement) {
				return fmt.Errorf("npc %d has invalid movement %q", id, mode.Movement)
			}
			if err := c.validateAttacks(mode.Attacks); err != nil {
				return fmt.Errorf("npc %d: %w", id, err)
			}
		}
	}

	for id, item := range c.items {
		if int(item.ID) != id {
			return fmt.Errorf("item %d has id %d", id, item.ID)
		}
		if err := c.validateAttacks(item.Attacks); err != nil {
			return fmt.Errorf("item %d: %w", id, err)
		}
	}

	for id, projectile := range c.projectiles {
		if int(projectile.ID) != id {
			return fmt.Errorf("projectile %d has id %d", id, projectile.ID)
		}
	}
	for id, bomb := range c.bombs {
		if int(bomb.ID) != id {
			return fmt.Errorf("bomb %d has id %d", id, bomb.ID)
		}
	}

	for tableID, table := range c.loot {
		for _, loot := range table.Entries {
			if int(loot.Item) >= len(c.items) {
				return fmt.Errorf("loot table %d has invalid item %d", tableID, loot.Item)
			}
		}
	}

	for collectionID, collection := range c.spawns {
		for _, spawn := range collection.Spawns {
			for _, npc := range spawn.NPCs {
				if int(npc.ID) >= len(c.npcs) {
					return fmt.Errorf("spawn collection %d has invalid npc %d", collectionID, npc.ID)
				}
			}
		}
	}

	if len(c.spawns) < 5 {
		return fmt.Errorf("expected at least 5 spawn collections")
	}
	return nil
}

func (c *Catalog) validateAttacks(attacks []Attack) error {
	for _, attack := range attacks {
		for _, projectile := range attack.Projectiles {
			if int(projectile.ID) >= len(c.projectiles) {
				return fmt.Errorf("invalid projectile %d", projectile.ID)
			}
		}
		for _, bomb := range attack.Bombs {
			if int(bomb.ID) >= len(c.bombs) {
				return fmt.Errorf("invalid bomb %d", bomb.ID)
			}
		}
		for _, summon := range attack.Summons {
			if int(summon.ID) >= len(c.npcs) {
				return fmt.Errorf("invalid summon %d", summon.ID)
			}
		}
	}
	return nil
}

func validMovement(movement string) bool {
	switch movement {
	case "wander", "chase", "run", "overshoot", "hover", "turret":
		return true
	default:
		return false
	}
}
