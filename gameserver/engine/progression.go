package engine

import "math"

// Keep the database's generated level expression in sync with this formula.
const ExpPerLevel int64 = 100

func LevelFromExp(exp int64) int64 { return 1 + max(exp, 0)/ExpPerLevel }

func (character *Character) AwardExp(amount int64) {
	if amount <= 0 || character.Dead || character.disconnected {
		return
	}
	character.Exp += min(amount, math.MaxInt64-character.Exp)
	character.Level = LevelFromExp(character.Exp)
	character.Send(character.PackFull(HANDSHAKE))
}
