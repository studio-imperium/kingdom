package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"gateway/gameservers"
	sessions "gateway/user_sessions"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	var database_error *pq.Error
	switch {
	case errors.Is(err, redis.Nil), errors.Is(err, sessions.ErrInvalidCredentials), errors.Is(err, sql.ErrNoRows):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials or session"})
	case errors.Is(err, sessions.ErrNameRequired), errors.Is(err, sessions.ErrCharacterLimit):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.As(err, &database_error) && database_error.Code == "23505":
		writeJSON(w, http.StatusConflict, map[string]string{"error": "record already exists"})
	default:
		log.Print(err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func readToken(w http.ResponseWriter, r *http.Request) (sessions.SessionToken, bool) {
	var token sessions.SessionToken
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && len(parts[1]) == hex.EncodedLen(len(token)) {
		if _, err := hex.Decode(token[:], []byte(parts[1])); err == nil {
			return token, true
		}
	}
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "expected Authorization: Bearer <64-character hex token>"})
	return token, false
}

func auth(register bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected a JSON object"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected one JSON object"})
			return
		}
		input.Name, input.Email = strings.TrimSpace(input.Name), strings.TrimSpace(input.Email)
		if input.Email == "" || input.Password == "" || (register && input.Name == "") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required; registration also requires name"})
			return
		}
		var token sessions.SessionToken
		var err error
		status := http.StatusOK
		if register {
			token, err = sessions.Signup(input.Name, input.Email, input.Password)
			status = http.StatusCreated
		} else {
			token, err = sessions.Login(input.Email, input.Password)
		}
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, status, map[string]string{"token": hex.EncodeToString(token[:])})
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	token, ok := readToken(w, r)
	if !ok {
		return
	}
	if err := sessions.Logout(token); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func create_character(w http.ResponseWriter, r *http.Request) {
	token, ok := readToken(w, r)
	if !ok {
		return
	}
	if err := sessions.NewCharacter(token); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"created": true})
}

func get_gameservers(w http.ResponseWriter, r *http.Request) {
	servers, err := gameservers.GetGameservers()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

func verify_token(w http.ResponseWriter, r *http.Request) {
	token, ok := readToken(w, r)
	if !ok {
		return
	}
	email, err := gameservers.ValidSession(token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"valid": true, "email": email})
}
