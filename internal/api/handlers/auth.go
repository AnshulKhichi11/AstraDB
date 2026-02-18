package handlers

import (
	"encoding/json"
	"net/http"

	"testDB/internal/auth"
)

type AuthHandlers struct {
	auth *auth.Auth
}

func NewAuthHandlers(a *auth.Auth) *AuthHandlers {
	return &AuthHandlers{auth: a}
}

// Login authenticates user
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON"})
		return
	}

	user, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		writeJSON(w, 401, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{
		"success": true,
		"apiKey":  user.APIKey,
		"role":    user.Role,
		"username": user.Username,
	})
}

// CreateUser creates new user (admin only)
func (h *AuthHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		Username string     `json:"username"`
		Password string     `json:"password"`
		Role     auth.Role  `json:"role"`
	}

	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON"})
		return
	}

	// Validate role
	if req.Role != auth.RoleAdmin && req.Role != auth.RoleReadWrite && req.Role != auth.RoleReadOnly {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid role"})
		return
	}

	user, err := h.auth.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{
		"success": true,
		"user": map[string]any{
			"username": user.Username,
			"role":     user.Role,
			"apiKey":   user.APIKey,
		},
	})
}

// ListUsers lists all users (admin only)
func (h *AuthHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users := h.auth.ListUsers()

	// Don't expose passwords
	userList := make([]map[string]any, 0, len(users))
	for _, u := range users {
		userList = append(userList, map[string]any{
			"username":  u.Username,
			"role":      u.Role,
			"active":    u.Active,
			"createdAt": u.CreatedAt,
		})
	}

	writeJSON(w, 200, map[string]any{
		"success": true,
		"users":   userList,
	})
}

// RotateAPIKey generates new API key
func (h *AuthHandlers) RotateAPIKey(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, 401, map[string]any{"success": false, "error": "Unauthorized"})
		return
	}

	newKey, err := h.auth.RotateAPIKey(user.Username)
	if err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{
		"success": true,
		"apiKey":  newKey,
	})
}

// DeleteUser deletes a user (admin only)
func (h *AuthHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		writeJSON(w, 400, map[string]any{"success": false, "error": "username required"})
		return
	}

	if err := h.auth.DeleteUser(username); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{
		"success": true,
		"message": "User deleted",
	})
}

// ChangePassword changes user password
func (h *AuthHandlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON"})
		return
	}

	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, 401, map[string]any{"success": false, "error": "Unauthorized"})
		return
	}

	if err := h.auth.ChangePassword(user.Username, req.OldPassword, req.NewPassword); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{
		"success": true,
		"message": "Password changed",
	})
}