package Handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	Constants "AuthServer/constants"
	Models "AuthServer/models"
	Utils "AuthServer/utils"
)

type Env struct {
	DB *sql.DB
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	IsLocked  bool   `json:"is_locked"`
	CreatedAt string `json:"created_at"`
}

type CredentialsReq struct {
	Username string `json:"username" example:"admin123"`
	Password string `json:"password" example:"password123"`
}

type UpdatePasswordReq struct {
	Password string `json:"password" example:"newpass123"`
}

// Signup godoc
// @Summary Create a new user
// @Description Register a new user with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body CredentialsReq true "User credentials"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/signup [post]
func (env *Env) Signup(w http.ResponseWriter, r *http.Request) {
	var req CredentialsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := Utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	stmt, err := env.DB.Prepare("INSERT INTO users(username, password) VALUES(?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(req.Username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"id":      id,
	})
}

// Signin godoc
// @Summary Login
// @Description Login with username and password to get JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body CredentialsReq true "User credentials"
// @Success 200 {object} map[string]string
// @Router /api/auth/signin [post]
func (env *Env) Signin(w http.ResponseWriter, r *http.Request) {
	var req CredentialsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	var hash string
	err := env.DB.QueryRow("SELECT id, username, password, is_locked FROM users WHERE username = ?", req.Username).Scan(&user.ID, &user.Username, &hash, &user.IsLocked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.IsLocked {
		http.Error(w, "Account is locked", http.StatusForbidden)
		return
	}

	if !Utils.CheckPasswordHash(req.Password, hash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := Utils.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

// GetUserInfo godoc
// @Summary Get User Info
// @Description Get current user information base on JWT
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} User
// @Router /api/auth/user [get]
func (env *Env) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(Constants.USER_DATA).(Models.Context)

	var user User
	err := env.DB.QueryRow("SELECT id, username, is_locked, created_at FROM users WHERE id = ?", userData.UserID).Scan(&user.ID, &user.Username, &user.IsLocked, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// UpdateUserInfo godoc
// @Summary Update User Password
// @Description Update the current user password
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdatePasswordReq true "New password"
// @Success 200 {object} map[string]string
// @Router /api/auth/user [put]
func (env *Env) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(Constants.USER_DATA).(Models.Context)

	var req UpdatePasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := Utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = env.DB.Exec("UPDATE users SET password = ? WHERE id = ?", hash, userData.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User updated successfully",
	})
}

// DeleteUser godoc
// @Summary Delete User
// @Description Delete the current user account
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/auth/user [delete]
func (env *Env) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(Constants.USER_DATA).(Models.Context)

	_, err := env.DB.Exec("DELETE FROM users WHERE id = ?", userData.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
}

// LockUser godoc
// @Summary Lock Account
// @Description Lock the current user account preventing further access
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/auth/lock [post]
func (env *Env) LockUser(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(Constants.USER_DATA).(Models.Context)

	_, err := env.DB.Exec("UPDATE users SET is_locked = 1 WHERE id = ?", userData.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User locked successfully",
	})
}

// RefreshToken godoc
// @Summary Refresh Token
// @Description Obtain a new JWT token using the current active one
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/auth/refresh [post]
func (env *Env) RefreshToken(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(Constants.USER_DATA).(Models.Context)

	token, err := Utils.GenerateToken(userData.UserID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
