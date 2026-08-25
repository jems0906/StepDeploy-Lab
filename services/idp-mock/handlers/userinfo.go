package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/example/stepdeploy-lab/shared/auth"
)

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("UserInfo request received from %s", r.RemoteAddr)

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logger.Warn("UserInfo request without authorization header")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userInfo := auth.UserInfo{
		Sub:   "user-123",
		Email: "user@example.com",
		Name:  "Test User",
	}

	logger.Info("UserInfo returned for token: %s", authHeader)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}
