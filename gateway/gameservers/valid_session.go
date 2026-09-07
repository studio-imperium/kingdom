package gameservers

import sessions "gateway/user_sessions"

func ValidSession(token sessions.SessionToken) (string, error) {
	return sessions.CheckToken(token)
}
