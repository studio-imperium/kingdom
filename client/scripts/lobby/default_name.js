function set_default_name() {
  const adjectives = [
    "Bumbling", "Busy", "Flustered", "Baffled", "Bashful", "Bouncy",
    "Clumsy", "Dawdling", "Dizzy", "Drowsy", "Fidgety", "Frazzled",
    "Giggly", "Goofy", "Grumpy", "Huffy", "Jolly", "Loopy",
    "Peckish", "Perplexed", "Puzzled", "Sleepy", "Snazzy", "Snoozy",
    "Soggy", "Spiffy", "Squirmy", "Startled", "Wobbly", "Wonky",
  ]
  // Enemy display names from gameserver/engine/assets/npcs.json.
  const enemies = [
    "Pirate", "Pirate Captain", "Nature Druid", "Nature Spirit",
    "Goblin Archer", "Goblin Warrior", "Goblin Cannon", "Goblin Chief",
    "Goblin Warlord", "Sand Golem", "Sand Spirit", "Sand Worm",
    "Sand Tribesman", "Sand Warrior", "Sand Brute", "Blowpiper",
    "Sand Caster", "Sand Cultist", "Dune Wraith",
  ]
  const pick = (names) => names[Math.floor(Math.random() * names.length)]
  const name = localStorage.getItem("kingdom.player_name") || sessionStorage.getItem("kingdom.player_name") || `${pick(adjectives)} ${pick(enemies)}`.replaceAll(" ", "")
  document.querySelector(".text_input").value = name
  localStorage.setItem("kingdom.player_name", name)
}
set_default_name()
document.querySelector(".text_input").addEventListener("input", event => localStorage.setItem("kingdom.player_name", event.target.value.trim()))
