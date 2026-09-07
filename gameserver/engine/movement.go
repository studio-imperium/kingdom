package engine

import (
	"time"
)

func sign(n float32) float32 {
	if n > 0 {
		return 1
	} else {
		return -1
	}
}

func (npc *Npc) Chase(delta time.Duration) {
	delta_time := float32(delta) / float32(time.Second)
	speed := float32(npcData[npc.id].Modes[npc.mode].Speed)
	target := NearbyPoint(npc.target, float32(Distance(npc, npc.target)))

	dx := sign(target.GetX() - npc.x)
	dy := sign(target.GetY() - npc.y)

	npc.x += dx * speed * float32(delta_time)
	npc.y += dy * speed * float32(delta_time)
}

func (npc *Npc) Run(delta time.Duration) {
	delta_time := float32(delta) / float32(time.Second)
	speed := float32(npcData[npc.id].Modes[npc.mode].Speed)
	target := NearbyPoint(npc.target, float32(Distance(npc, npc.target)))

	dx := sign(npc.x - target.GetX())
	dy := sign(npc.y - target.GetY())

	npc.x += dx * speed * delta_time
	npc.y += dy * speed * delta_time
}

func (npc *Npc) Hovering() bool {
	dist := float32(Distance(npc, npc.target))
	too_close := (npcData[npc.id].Range / 2) - 1
	too_far := (npcData[npc.id].Range / 2) + 1

	return (dist < too_far &&
		dist > too_close)
}

func (npc *Npc) Hover(delta time.Duration) {
	dist := float32(Distance(npc, npc.target))
	too_close := (npcData[npc.id].Range / 2) - 1
	too_far := (npcData[npc.id].Range / 2) + 1

	if dist > too_far {
		npc.Chase(delta)
		npc.Look(nil)
	} else if dist < too_close {
		npc.Run(delta)
		npc.Look(nil)
	} else if entity, ok := npc.target.(Entity); ok {
		npc.Look(entity)
	}
}

func (npc *Npc) Turret(delta time.Duration) {
	if entity, ok := npc.target.(Entity); ok {
		npc.Look(entity)
	}
}

func (npc *Npc) Overshoot(delta time.Duration) {
	delta_time := float32(delta) / float32(time.Second)
	speed := float32(npcData[npc.id].Modes[npc.mode].Speed)
	target := NearbyPoint(npc.target, 1)

	dx := sign(target.GetX() - npc.x)
	dy := sign(target.GetY() - npc.y)

	npc.x += dx * speed * delta_time
	npc.y += dy * speed * delta_time
}

func (npc *Npc) Wander(delta time.Duration) {
	delta_time := float32(delta) / float32(time.Second)
	speed := float32(npcData[npc.id].Modes[npc.mode].Speed)
	target := npc.target

	if target == nil || Distance(target, npc) < 1 {
		npc.target = NearbyPoint(npc.origin, 4)
		target = npc.target
	}

	dx := sign(target.GetX() - npc.x)
	dy := sign(target.GetY() - npc.y)

	npc.x += dx * speed * delta_time
	npc.y += dy * speed * delta_time
}

func (npc *Npc) Travel(delta time.Duration) {
	delta_time := float32(delta) / float32(time.Second)
	speed := float32(npcData[npc.id].Modes[npc.mode].Speed)
	target := npc.target

	if target == nil || Distance(target, npc) < 1 {
		npc.target = NearbyPoint(npc.origin, 16)
		target = npc.target
	}

	dx := sign(target.GetX() - npc.x)
	dy := sign(target.GetY() - npc.y)

	npc.x += dx * speed * delta_time
	npc.y += dy * speed * delta_time
}
