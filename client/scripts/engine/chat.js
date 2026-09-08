const message_box = document.getElementById("message_box")
const messages_container = document.getElementById("messages")
let chat_focused = false

message_box.addEventListener("focusin", () => (chat_focused = true))
message_box.addEventListener("focusout", () => (chat_focused = false))

document.addEventListener("keydown", (e) => {
  if (e.key == "Enter" && !chat_focused) {
    message_box.focus()
    e.preventDefault()
  } else if (e.key == "Enter") {
    submit_message(e)
  }

  if (e.key == "/" && !chat_focused) {
    message_box.focus()
    message_box.value = "/"
    e.preventDefault()
  }
})

function add_message(id, msg, sender) {
  let message = document.createElement("div")

  if (id == 0) {
    message.className = "message error"
  } else if (id == 1) {
    message.className = "message success"
  } else {
    message.className = "message"
  }

  let label = document.createElement("label")
  label.className = "outlined_text"
  label.textContent = sender

  let contents = document.createElement("p")
  contents.className = "outlined_text"
  contents.textContent = msg

  if (id > 1) {
    message.appendChild(label)
  }
  message.appendChild(contents)
  messages_container.appendChild(message)
}

function submit_message(e) {
  e.preventDefault()

  send_message(message_box.value)

  message_box.value = ""
  message_box.blur()
}
