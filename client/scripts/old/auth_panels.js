function update_account_buttons() {
  const logged_in = account_session?.valid && !account_session.guest
  document.getElementById("login_button").classList.toggle("hidden", !!logged_in)
  document.getElementById("logout_button").classList.toggle("hidden", !logged_in)
  update_lobby_character()
}

async function submit_account(event, action) {
  event.preventDefault()
  const form = event.currentTarget, error = form.querySelector('[role="alert"]')
  const confirm = form.elements.password_confirmation
  if (confirm && confirm.value !== form.elements.password.value) {
    confirm.setCustomValidity("Passwords do not match.")
    confirm.reportValidity()
    return
  }
  const button = form.querySelector('[type="submit"]')
  button.disabled = true
  error.classList.add("hidden")
  try {
    if (action === "logout") {
      await logout()
      set_default_name()
    }
    else {
      const session = await (action === "register" ? register : login)(form.elements.email.value.trim(), form.elements.password.value)
      if (!session?.valid || session.guest) throw new Error("Unable to restore your account session.")
    }
    form.reset()
    switch_screen("default")
  } catch (failure) {
    error.textContent = failure.message
    error.classList.remove("hidden")
  } finally {
    button.disabled = false
    update_account_buttons()
  }
}

account_ready.then(update_account_buttons)
