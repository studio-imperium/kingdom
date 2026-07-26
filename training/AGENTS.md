# training/ — agent notes

Notes from the session that built this directory (2026-07-01). This is about
*how to work here*, not *what's here* — see `README.md` for the architecture,
observation/action/reward spec, and file-by-file mapping back to the Go
server. If this file and the code ever disagree, trust the code.

## Context

`training/` is a from-scratch C++ port of the gameplay-relevant parts of
`server/engine` (the Go game server), built to train an RL policy that plays
the game the way a single player would. Not networked, not multi-agent yet.

## Conventions established this session

**Port faithfully — don't "fix" the source.** When porting Go logic, match
it exactly, including things that look like quirks or bugs (e.g. a
non-piercing projectile can hit more than one NPC in the same tick, because
Go's hit-collection is deferred; the NPC auto-aim angle formula has an
unintuitive `-90` offset). Cite the exact Go file/function being mirrored.
Only deviate when the user explicitly asks for different behavior, and say
so plainly when you do.

**Verify against actual source before asserting anything load-bearing.**
Subagent research summaries are a starting point, not ground truth. This
session, two specific claims from subagent reports (where hit-resolution
actually lives; whether player movement is server-clamped) were re-verified
by directly reading the Go source before being presented to the user as
fact. Do this before stating anything specific enough that being wrong would
mislead a design decision.

**Flag every placeholder as a placeholder, in code comments and README —
don't let a made-up number read as tuned.** Reward weights, arena tier
distances/populations, the gear-score formula, observation entity caps, the
bomb-throw-distance constant: all invented for this project (no real
equivalent existed to port), and all explicitly marked as starting points
that need real training data before they mean anything.

**Reward shaping needs an exploitability check, not just "does this reward
the right event."** First pass at the gear-upgrade reward compared against
the immediately-prior gear score, which was farmable (equip good item →
paid; unequip → free; re-equip same item → paid again, forever). Fixed by
tracking a per-episode high-water mark, only paying out on a new best.
Apply the same scrutiny to any new reward term: can the agent repeat a cheap
action to collect it indefinitely without genuine progress?

**When a design choice has real strategic depth, don't auto-resolve it
away.** Attack targeting was originally auto-aim-at-nearest-enemy; changed
to a policy-controlled direction because auto-targeting forecloses
prioritization (finishing off a low-health target, avoiding a dangerous
one). Default to giving the policy control plus the observation data to
decide, over hard-coding the "obviously correct" heuristic.

**Use TaskCreate/TaskUpdate for multi-file build work here.** Sessions in
this directory tend to be large, multi-step ports/builds; task tracking has
been used throughout and should continue.

## Current status

- Environment builds clean and runs:
  `cmake -S . -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build -j 8 && ./build/rollout_demo`
- `Observation`/`Action` are fixed-size and trivially copyable (enforced by
  `static_assert`, not just documented), specifically so they're ready for
  PufferLib's shared-memory `PufferEnv` convention.
- Decided direction: train via PufferLib (native C/C++ env, their PPO
  trainer, their vectorization) instead of generic pybind11 + SB3/CleanRL.
  **Not yet built:** the actual `PufferEnv` `.h`/`.c` wrapper and `.ini`
  config -- `Observation`/`Action` are ready for it, nothing consumes them
  from Python yet.
- **Not built yet:** any way to visually verify the sim (trace/replay/
  visualizer). Validated so far only via aggregate stats from random
  rollouts in `rollout_demo` -- nobody has watched an episode.
- Multi-agent/self-play is explicitly deferred per the user -- `World` owns
  exactly one `Character`, not a list of them.
