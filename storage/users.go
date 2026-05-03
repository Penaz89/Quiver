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
	Role         string `json:"role,omitempty"`
	MustChange   bool   `json:"must_change,omitempty"`
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

// CheckUserAuth checks if a username and password match. Returns the User and true if matched.
func CheckUserAuth(dataDir, username, password string) (*User, bool) {
	users, err := LoadUsers(dataDir)
	if err != nil {
		return nil, false
	}
	for _, u := range users {
		if u.Username == username {
			err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
			if err == nil {
				return &u, true
			}
			return nil, false
		}
	}
	return nil, false
}

// GetUser gets a user by username
func GetUser(dataDir, username string) (*User, error) {
	users, err := LoadUsers(dataDir)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, os.ErrNotExist
}

// DeleteUser deletes a user by username
func DeleteUser(dataDir, username string) error {
	users, err := LoadUsers(dataDir)
	if err != nil {
		return err
	}
	var newUsers []User
	found := false
	for _, u := range users {
		if u.Username != username {
			newUsers = append(newUsers, u)
		} else {
			found = true
		}
	}
	if !found {
		return os.ErrNotExist
	}
	
	// Try to remove their directory
	userDir := filepath.Join(dataDir, "users", username)
	_ = os.RemoveAll(userDir)
	
	return SaveUsers(dataDir, newUsers)
}

// UpdateUserPassword updates the password for a user. If mustChange is false, clears the MustChange flag.
func UpdateUserPassword(dataDir, username, password string, mustChange bool) error {
	users, err := LoadUsers(dataDir)
	if err != nil {
		return err
	}
	
	found := false
	for i, u := range users {
		if u.Username == username {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			users[i].PasswordHash = string(hash)
			users[i].MustChange = mustChange
			found = true
			break
		}
	}
	
	if !found {
		return os.ErrNotExist
	}
	
	return SaveUsers(dataDir, users)
}

// EnsureAdminUser creates the default admin user if it doesn't exist.
func EnsureAdminUser(dataDir string) error {
	users, err := LoadUsers(dataDir)
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Username == "admin" {
			return nil
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	users = append(users, User{
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         "admin",
		MustChange:   true,
	})
	
	userDir := filepath.Join(dataDir, "users", "admin")
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}
	
	return SaveUsers(dataDir, users)
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
