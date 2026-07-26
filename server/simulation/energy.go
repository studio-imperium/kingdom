package simulation

type EnergyConfig struct {
	Maximum        float32
	Regeneration   float32
	MovementCost   float32
	ProjectileCost float32
	PowerCost      float32
}

type Energy struct {
	Amount   float32
	Config   EnergyConfig
	movement Movement
}

func NewEnergy(config EnergyConfig, amount float32) *Energy {
	config = config.Sanitize()
	return &Energy{
		Amount: min(clampFinite(amount), config.Maximum),
		Config: config,
	}
}

func (c EnergyConfig) Sanitize() EnergyConfig {
	c.Maximum = max(clampFinite(c.Maximum), 0)
	c.Regeneration = max(clampFinite(c.Regeneration), 0)
	c.MovementCost = max(clampFinite(c.MovementCost), 0)
	c.ProjectileCost = max(clampFinite(c.ProjectileCost), 0)
	c.PowerCost = max(clampFinite(c.PowerCost), 0)
	return c
}

func (c EnergyConfig) Cost(action Action, seconds float32) float32 {
	action = action.Sanitize()
	seconds = max(clampFinite(seconds), 0)
	c = c.Sanitize()

	cost := action.Movement.Sum() * seconds * c.MovementCost
	for index := range int(action.Count) {
		cost += c.ProjectileCost
		cost += action.Projectiles[index].Power * c.PowerCost
	}
	return cost
}

func (e *Energy) Step(seconds float32, think func() Action) (Action, bool) {
	seconds = max(clampFinite(seconds), 0)
	e.Config = e.Config.Sanitize()
	e.Amount = min(clampFinite(e.Amount)+e.Config.Regeneration*seconds, e.Config.Maximum)

	if e.Amount < 0 {
		e.movement = e.movement.Half()
		return Action{Movement: e.movement}, false
	}

	action := think().Sanitize()
	e.Amount -= e.Config.Cost(action, seconds)
	e.movement = action.Movement
	return action, true
}

func clampFinite(value float32) float32 {
	if !finite(value) {
		return 0
	}
	return value
}
