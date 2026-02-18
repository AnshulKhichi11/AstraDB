package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleReadWrite Role = "readWrite"
	RoleReadOnly  Role = "readOnly"
)

type User struct {
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Never expose in JSON
	Role      Role      `json:"role"`
	APIKey    string    `json:"apiKey"`
	CreatedAt time.Time `json:"createdAt"`
	Active    bool      `json:"active"`
}

type Auth struct {
	users  map[string]*User      // username -> user
	apiKeys map[string]*User     // apiKey -> user
	mu     sync.RWMutex
}

func New() *Auth {
	auth := &Auth{
		users:   make(map[string]*User),
		apiKeys: make(map[string]*User),
	}

	// Create default admin user
	defaultAdmin := &User{
		Username:  "admin",
		Password:  hashPassword("admin123"), // CHANGE IN PRODUCTION!
		Role:      RoleAdmin,
		APIKey:    generateAPIKey(),
		CreatedAt: time.Now(),
		Active:    true,
	}

	auth.users[defaultAdmin.Username] = defaultAdmin
	auth.apiKeys[defaultAdmin.APIKey] = defaultAdmin

	return auth
}

// CreateUser creates a new user
func (a *Auth) CreateUser(username, password string, role Role) (*User, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.users[username]; exists {
		return nil, errors.New("user already exists")
	}

	user := &User{
		Username:  username,
		Password:  hashPassword(password),
		Role:      role,
		APIKey:    generateAPIKey(),
		CreatedAt: time.Now(),
		Active:    true,
	}

	a.users[username] = user
	a.apiKeys[user.APIKey] = user

	return user, nil
}

// Login authenticates user and returns API key
func (a *Auth) Login(username, password string) (*User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	user, exists := a.users[username]
	if !exists {
		return nil, errors.New("invalid credentials")
	}

	if !user.Active {
		return nil, errors.New("user is disabled")
	}

	if !checkPassword(user.Password, password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// Verify checks if API key is valid
func (a *Auth) Verify(apiKey string) (*User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	user, exists := a.apiKeys[apiKey]
	if !exists {
		return nil, errors.New("invalid API key")
	}

	if !user.Active {
		return nil, errors.New("user is disabled")
	}

	return user, nil
}

// RotateAPIKey generates new API key for user
func (a *Auth) RotateAPIKey(username string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	user, exists := a.users[username]
	if !exists {
		return "", errors.New("user not found")
	}

	// Remove old API key
	delete(a.apiKeys, user.APIKey)

	// Generate new API key
	newKey := generateAPIKey()
	user.APIKey = newKey
	a.apiKeys[newKey] = user

	return newKey, nil
}

// DeleteUser removes a user
func (a *Auth) DeleteUser(username string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	user, exists := a.users[username]
	if !exists {
		return errors.New("user not found")
	}

	delete(a.apiKeys, user.APIKey)
	delete(a.users, username)

	return nil
}

// ListUsers returns all users
func (a *Auth) ListUsers() []*User {
	a.mu.RLock()
	defer a.mu.RUnlock()

	users := make([]*User, 0, len(a.users))
	for _, user := range a.users {
		users = append(users, user)
	}

	return users
}

// ChangePassword updates user password
func (a *Auth) ChangePassword(username, oldPassword, newPassword string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	user, exists := a.users[username]
	if !exists {
		return errors.New("user not found")
	}

	if !checkPassword(user.Password, oldPassword) {
		return errors.New("invalid old password")
	}

	user.Password = hashPassword(newPassword)
	return nil
}

// Helper functions

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "astra_" + hex.EncodeToString(bytes)
}

// Check permissions
func (u *User) CanRead() bool {
	return u.Role == RoleAdmin || u.Role == RoleReadWrite || u.Role == RoleReadOnly
}

func (u *User) CanWrite() bool {
	return u.Role == RoleAdmin || u.Role == RoleReadWrite
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}