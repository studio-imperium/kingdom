# training/ -- single-agent C++ training environment

A from-scratch C++ port of the gameplay-relevant parts of `server/engine`,
for training an RL policy to play the game as a single player would. Not
networked, not multi-agent (yet) -- just a fast, headless, steppable copy of
the sim: `Reset()` / `Step(action) -> observation, reward, done`.

This is a first working draft. It builds and runs (see "Build & run" below),
but reward magnitudes, arena tuning, and a few movement decisions below are
explicitly called out as placeholders to revisit once you can actually watch
a policy train.

## Why a from-scratch port instead of reusing the Go engine

The real game's per-client "simulation" loop (`engine/simulation.go`) isn't
just network view-culling -- it's also where hit detection and loot pickup
actually happen (the main `Engine.Run()` tick only moves things). Networking,
per-client goroutines, mutexes, and channels are gone; single-threaded,
single-player, stepped explicitly instead of ticked against a wall clock.

The real game's procedural map (biome/cell spawning, `engine/map.go`,
`assets/spawns.json`) is also gone, replaced by a much simpler arena (see
below) -- there's no collision/tiles in the real game either, so nothing of
substance was lost by dropping the map system for training purposes.

## File map (Go source -> C++)

| Go source | C++ file(s) | Notes |
|---|---|---|
| `engine/assets.go` + `assets/*.json` | `assets.hpp/.cpp` | Same JSON files, loaded from `server/engine/assets` directly (see `KINGDOMS_ASSETS_DIR` in CMakeLists.txt) so balance data can't drift from the live game. Does NOT load `spawns.json`/`tiles.json`/`maps/` -- see `arena.hpp`. |
| `engine/character.go` | `character.hpp/.cpp` | Dropped: `send` channel, `angle`, `Pack`/`PackFull`. |
| `engine/npc.go` | `npc.hpp/.cpp` | Dropped: `nearby`/`EnterView`/`ExitView` (real per-client view culling), `looking` (cosmetic). `target` (a Go interface: character or wander point) collapses to `targeting_player: bool` + `movement_point: Vec2` since there's only one possible player. |
| `engine/projectile.go`, `bomb.go`, `loot.go` | `projectile.hpp/.cpp`, `bomb.hpp`, `loot.hpp` | Direct ports. |
| `engine/combat.go` | `combat.hpp/.cpp` | Hit-test math + spawn factories (see "Attack direction" below). |
| `engine/movement.go` | `movement.hpp/.cpp` | All 6 NPC movement modes, including the random-jitter aim point `Chase`/`Run`/`Overshoot` recompute every tick (`NearbyPoint`) -- this is real behavior in the live game, not a training-only simplification. |
| `engine/engine.go` (`Run()`) + `engine/simulation.go` (`StartSimulation()`) | `world.hpp/.cpp` | Merged into one `World::Step()`. See tick order below. |
| `engine/map.go`, `assets/spawns.json` | `arena.hpp/.cpp` | Replaced, not ported -- see below. |
| (none -- new) | `action.hpp`, `observation.hpp/.cpp`, `reward.hpp/.cpp`, `environment.hpp/.cpp` | The RL-facing layer. |

## What the policy sees (`Observation`, `observation.hpp`)

Per tick:
- Self: position, health/max_health, attack cooldown remaining (raw value,
  can go slightly negative -- see the combo-reset note in `character.hpp`),
  derived stats (speed/power/reload/regen), current hand slot, equipped
  head/body, and all 24 backpack slots (`-1` = empty, else item id).
- Every live NPC: position, remaining health, NPC type id.
- Every live projectile: position, damage, whether it's enemy-fired (i.e.
  threatens the player) or player-fired.
- Every live bomb: position, damage, blast radius, remaining airtime.
- Every live loot drop: position, item id.

No cosmetic facing angle anywhere -- confirmed not worth training. Attacks do
get an explicit direction, though (see "Attack direction" below), and every
NPC's remaining health is in here specifically so the policy can use that
direction to prioritize -- e.g. finish off something low on health -- instead
of the environment forcing a fixed targeting rule.

No render-distance culling: everything alive is included every tick (up to
the caps below). The real game's 16-unit cutoff existed to keep bandwidth/CPU
down across many clients; a training arena with a dozen NPCs doesn't need it.

**Fixed-size, not a variable-length list.** `Observation::entities` is a
`std::array<ObservedEntity, kMaxObservedEntities>`, not a `std::vector` --
slots `[0, kMaxNpcSlots)` are NPCs, then `kMaxProjectileSlots` worth of
projectiles, then bombs, then loot, always in that order. Whichever entities
are nearest to the player fill the slots for their type; anything past the
cap is dropped (farthest first) and counted in `truncated_count` rather than
silently missing. This is entirely about the eventual training backend --
PufferLib's native `PufferEnv` convention writes each environment's
observation directly into a fixed-size shared-memory block, so nothing
tensor-facing can be a ragged/variable-length container. Current caps (16
NPCs / 24 projectiles / 12 bombs / 12 loot -- see `observation.hpp`) are
sized off the arena's steady-state population and confirmed to never
truncate under random play in `rollout_demo` (`max_truncated` stays `0` even
across full-length 1000-simulated-second episodes) -- but they're still
placeholders, not something derived from real training data, since a policy
that actively seeks out fights could plausibly push projectile counts higher
than random movement does. `Action` (`action.hpp`) needed no equivalent
change -- it was already a fixed-size, trivially-copyable struct.

## What the policy controls (`Action`, `action.hpp`)

- `move: Vec2` -- desired direction; magnitude is clamped to 1 internally,
  then scaled by `speed * dt`. Same formula NPCs already use
  (`direction * speed * delta_time`) -- real player movement in the live
  game has no server-side speed enforcement at all (`Character.Move` just
  sets x/y directly from whatever the client sends), so this is *new* logic
  modeled on the NPC formula, not a port of existing player-movement code.
- `attack: bool` -- attempt to attack or use the held item (cooldown-gated).
  Consumables (e.g. potions) and weapons share this one trigger, dispatched
  by what's currently in `hand` -- this mirrors `characterAttack` in
  `clients.go` exactly (it returns immediately after `UseItem` for anything
  with `on_use` set), so no separate "use item" action was needed.
- `attack_direction: Vec2` -- only read when `attack=true`; see "Attack
  direction" below. Independent of `move`: the policy can strafe one way
  while firing another.
- `do_equip` + `equip_from`/`equip_to` -- backpack slots are `[0, 24)`, head
  is `24`, body is `25`. Invalid moves (e.g. equipping a body item into the
  head slot) are no-ops, matching `ChangeInventory`.
- `do_switch` + `switch_to_slot` -- changes which backpack slot is "active"
  (`SelectSlot`).

### Attack direction (policy-controlled, not auto-aimed)

Attacks are directional, and the direction is a policy decision, not
something the environment picks for it. The reasoning: auto-targeting
"nearest enemy" forecloses any notion of priority (e.g. finishing off a
low-health NPC that's a couple steps further away than the closest one, or
avoiding aiming at all when nothing worth hitting is close), so the policy
gets `attack_direction` and decides for itself, using the per-NPC health/
position already in the observation.

- **NPCs** are unaffected by this -- they still auto-aim at the player via
  `MakeAimedProjectile`/`MakeAimedBomb` in `combat.hpp`, which is how the
  original game worked too (NPC attacks were always auto-aimed, never
  client-driven). That angle formula is ported **verbatim** from
  `npc.Projectile()` in `combat.go`, `-90` offset and all.
- **Player attacks** use `MakeDirectedProjectile`/`MakeDirectedBomb`
  instead, which is actually the *more* faithful option here: it's the real
  `(projectile.Angle + angle) % 360` formula from `characterAttack` in
  `clients.go`, just with `angle` now coming from `attack_direction` instead
  of a client mouse position. Player-thrown bombs land
  `kPlayerBombThrowDistance` (placeholder, `types.hpp`) away in that
  direction -- there's no data-driven "throw range" to copy, since the real
  game let the client click anywhere on the ground.
- A near-zero `attack_direction` (i.e. no meaningful direction given) skips
  firing and is counted as a miss, same as an out-of-range or otherwise
  unresolved shot.

## Reward (`reward.hpp/.cpp`)

Per tick, from `TickEvents` gathered during `World::Step`. The weights below
are only the C++ defaults (`RewardWeights`, reward.hpp): training runs get
them from PufferLib's `config/kingdoms.ini` `[env]` section (`reward_*` keys),
overridable per run as CLI flags (e.g. `--env.reward-kill 10`) -- no rebuild
needed to tune them.

| Event | Weight (placeholder) |
|---|---|
| Damage dealt (per point) | `+1.0` |
| Damage taken (per point) | `-1.0` |
| Kill | `+5.0` |
| Loot picked up (per item) | `+2.0` |
| Gear upgrade (per point of `GearScore` improvement from an equip) | `+3.0` |
| Missed attack (fired/attempted but hit nothing) | `0.0` (disabled -- at `-0.5`, random-play misses accrued ~`-700`/episode, making early death outscore survival; revisit as shaping once a policy trains) |
| Death (terminal) | `-100.0` |

None of these are balanced numbers -- they're starting points sized so death
clearly dominates and misses are a minor nudge, per your spec. `GearScore`
(also `reward.hpp`) sums `(stat - 1.0)` across the 5 multiplier fields
(health/regen/speed/damage/reload) for equipped head + body -- all 5 are
"higher is better" for the player, so this is a plain additive placeholder,
not a balanced valuation. The "gear upgrade" shaping term isn't something
you asked for explicitly, but seemed like an obvious gap: picking loot up
and equipping it are different actions (`AddItemOrBust` vs `ChangeInventory`),
so without it the agent has no per-tick signal that equipping upgrades
(rather than hoarding them in the backpack) matters, only the terminal score.

**Gear upgrade is measured against the episode's best score so far, not the
previous instant.** `World` tracks `best_gear_score_`, updated in
`ResolvePlayerAction`, and only pays out when an equip pushes
`GearScore(head, body)` past that high-water mark. This matters: comparing
only against "whatever was worn a moment ago" is exploitable -- equip a good
item (paid), take it back off (free, no penalty), re-equip the same item
(paid again) is free reward on a loop with zero actual progress. Comparing
against the episode-best closes that off -- only a loadout genuinely better
than anything worn so far this episode ever pays out again.

## Episode (`Environment`, `environment.hpp/.cpp`)

- Ends immediately when the player's health drops below 1 (`done = true` on
  that `Step()` call), or after `max_ticks` steps (constructor arg, default
  20,000) -- a time-limit truncation so an agent that learns to avoid all
  danger still ends episodes. No death penalty on timeout.
- `Environment::FinalGearScore()` reports `World::best_gear_score()` -- the
  episode high-water mark described above, not just whatever happens to be
  equipped the instant the episode ends -- meant for optimizing "gear
  collected" as you described, at whatever cadence you want to log it
  (terminal reward, separate metric, curriculum signal, etc. -- not wired to
  anything yet).
- `Reset(seed)` reseeds the RNG, respawns the player at the origin with the
  default loadout (Kendo Stick + default head/body, matching
  `DefaultCharacter` in `character.go`), and repopulates the arena.

## Arena (`arena.hpp/.cpp`)

Replaces the real game's biome/cell spawn system with 3 concentric distance
bands around the origin (player spawn):

| Tier | Distance band | Target population |
|---|---|---|
| Near (easy) | `[0, 40)` | 4 |
| Mid | `[40, 100)` | 4 |
| Far (hard) | `[100, 200)` | 4 |

NPC types are bucketed into tiers by a `difficulty_score` computed once at
asset-load time (`health + range + 20 * hardest single hit it can land`) --
sorted ascending and split into thirds, so tiering stays in sync
automatically if `npcs.json` changes rather than needing a hand-maintained
list. NPCs with no `modes` are excluded from spawning entirely (they can't
move or attack, and would index out of bounds trying).

Every 5 simulated seconds, each tier below its target population gets topped
up. All of the numbers in this section (band distances, population targets,
the respawn check interval, the difficulty formula's weights) are starting
placeholders -- there was no real equivalent to port, since the live game's
difficulty curve comes from the biome map, which training doesn't model.

## Known simplifications worth knowing about

- **NPC targeting/aggro** collapses Go's "nearest of N characters" to a
  single distance check, since there's only one player. Faithful for
  single-agent; will need revisiting for multi-agent/self-play.
- **Soulbound loot** still checks the "dealt enough damage personally"
  threshold against the one player, same as the real game.
- **Entity ids** are an incrementing counter, not `rand.Uint32()` -- simpler,
  and makes a `World` fully reproducible from its seed. Gameplay never
  depended on ids being random, only unique.
- **A non-piercing projectile can still hit more than one NPC in the same
  tick** if their hitboxes overlap at that exact instant. This looks like a
  quirk of the original Go code (hits are collected, then applied, so
  `Dead` isn't set in time to stop a second hit that same tick) rather than
  intentional design, but it's kept for fidelity rather than "fixed."

## Training backend: leaning toward PufferLib

Decided direction (not yet built): train against
[PufferLib](https://github.com/PufferAI/PufferLib) rather than a generic
pybind11 + Stable-Baselines3/CleanRL setup. Its native `PufferEnv`
convention is built around exactly this shape of project -- a fast custom
C/C++ simulation that writes its observation directly into a shared-memory
buffer PufferLib hands it, instead of paying a marshaling cost every step.
Their own bundled example environments ("Ocean") are written the same way.
It also ships its own PPO-based trainer (PuffeRL) and fast shared-memory
vectorization, so this one decision covers three separate concerns (Python
bindings, training algorithm, and running many envs in parallel) rather than
choosing each independently.

The `Observation`/`Action` redesign above (fixed-size, trivially copyable)
was the prerequisite step for this -- PufferLib's shared-memory model needs
fixed-size blocks per environment instance, which a variable-length entity
list can't provide. What's still ahead: the actual `PufferEnv` integration
(a `.h`/`.c` wrapper following their convention, `.ini` config -- their
"Squared" single-agent example is the suggested template to start from) and
picking concrete network architecture for the entity list (padding-aware,
so the policy doesn't learn spurious patterns from `ObservedEntityType::None`
slots).

Worth knowing before committing further: PufferLib is a smaller, more
specialized project than Stable-Baselines3/RLlib -- less ecosystem/community
depth if the custom integration proves annoying. Its docs also didn't
clearly cover Dict/variable-length observation handling as of this writing,
so the entity-list-to-tensor design above is this project's own choice, not
something copied from their docs.

## Not done yet

- The actual PufferEnv integration -- see above. Observation/Action are
  ready for it; the wrapper itself isn't written.
- Any actual policy/training code -- this is the environment only.
- Multi-agent/self-play -- explicitly out of scope for now per your
  original message; `World` currently owns exactly one `Character`, not a
  list, since guessing at the eventual multi-agent API shape now seemed
  more likely to be wrong than useful.
- Reward/arena/gear-score tuning -- everything numeric above is a
  placeholder, not a balanced first pass at "good" values.
- A way to actually watch an episode (trace/replay/visualizer) -- still only
  validated via aggregate stats from random rollouts, not visual confirmation
  that movement/combat/aiming look right.

## Build & run

```sh
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j 8
./build/rollout_demo
```

`rollout_demo` (`apps/rollout_demo.cpp`) runs 5 random-action episodes and
prints per-episode stats (reward, damage dealt/taken, kills, loot, final
gear score) -- it's a smoke test proving the pipeline runs end to end, not
a real baseline (it only randomizes movement and attack/attack_direction,
not equip/switch).
