function random_loading_message() {
  const messages = [
    "Press [E] to open your backpack.",
    "Leather armor makes you run faster.",
    "Bows can pierce multiple enemies.",
    "Bombs can be thrown from any range.",
  ]
  return "Did you know: " + messages[Math.floor(Math.random() * messages.length)]
}
