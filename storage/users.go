package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

// User represents a registered user.
type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

// GetUsersFile returns the path to the users registry.
func GetUsersFile(dataDir string) string {
	return filepath.Join(dataDir, "users.json")
}

// LoadUsers reads the users registry.
func LoadUsers(dataDir string) ([]User, error) {
	path := GetUsersFile(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// SaveUsers writes the users registry.
func SaveUsers(dataDir string, users []User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetUsersFile(dataDir), data, 0644)
}

// CheckUserAuth checks if a username and password match.
func CheckUserAuth(dataDir, username, password string) bool {
	users, err := LoadUsers(dataDir)
	if err != nil {
		return false
	}
	for _, u := range users {
		if u.Username == username {
			err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
			return err == nil
		}
	}
	return false
}

// CreateUser creates a new user if the username doesn't exist.
func CreateUser(dataDir, username, password string) error {
	users, err := LoadUsers(dataDir)
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Username == username {
			return os.ErrExist
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	users = append(users, User{
		Username:     username,
		PasswordHash: string(hash),
	})
	
	// Create user's personal data directory
	userDir := filepath.Join(dataDir, "users", username)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}
	
	return SaveUsers(dataDir, users)
}

// GetUserDir returns the personal data directory for a user.
func GetUserDir(dataDir, username string) string {
	return filepath.Join(dataDir, "users", username)
}
