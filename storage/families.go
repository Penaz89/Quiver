package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// Family represents a shared workspace.
type Family struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"` // Usernames
}

// GetFamiliesFile returns the path to the families registry.
func GetFamiliesFile(dataDir string) string {
	return filepath.Join(dataDir, "families.json")
}

// LoadFamilies reads the families registry.
func LoadFamilies(dataDir string) ([]Family, error) {
	path := GetFamiliesFile(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Family{}, nil
		}
		return nil, err
	}
	var families []Family
	if err := json.Unmarshal(data, &families); err != nil {
		return nil, err
	}
	return families, nil
}

// SaveFamilies writes the families registry.
func SaveFamilies(dataDir string, families []Family) error {
	data, err := json.MarshalIndent(families, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetFamiliesFile(dataDir), data, 0644)
}

// generateFamilyID generates a random hex string for family ID.
func generateFamilyID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateFamily creates a new family.
func CreateFamily(dataDir, name, creatorUsername string) (Family, error) {
	families, err := LoadFamilies(dataDir)
	if err != nil {
		return Family{}, err
	}

	family := Family{
		ID:      generateFamilyID(),
		Name:    name,
		Members: []string{creatorUsername},
	}

	families = append(families, family)

	// Create family's data directory
	familyDir := GetFamilyDir(dataDir, family.ID)
	if err := os.MkdirAll(familyDir, 0700); err != nil {
		return Family{}, err
	}

	return family, SaveFamilies(dataDir, families)
}

// GetFamilyDir returns the data directory for a family.
func GetFamilyDir(dataDir, familyID string) string {
	return filepath.Join(dataDir, "families", familyID)
}

// GetUserFamilies returns all families a user is a member of.
func GetUserFamilies(dataDir, username string) ([]Family, error) {
	families, err := LoadFamilies(dataDir)
	if err != nil {
		return nil, err
	}

	var userFamilies []Family
	for _, f := range families {
		for _, m := range f.Members {
			if m == username {
				userFamilies = append(userFamilies, f)
				break
			}
		}
	}
	return userFamilies, nil
}

// AddMemberToFamily adds a user to an existing family.
func AddMemberToFamily(dataDir, familyID, username string) error {
	families, err := LoadFamilies(dataDir)
	if err != nil {
		return err
	}

	// Verify user exists
	_, err = GetUser(dataDir, username)
	if err != nil {
		return err // User does not exist
	}

	found := false
	for i, f := range families {
		if f.ID == familyID {
			// Check if already a member
			for _, m := range f.Members {
				if m == username {
					return nil // Already a member
				}
			}
			families[i].Members = append(families[i].Members, username)
			found = true
			break
		}
	}

	if !found {
		return os.ErrNotExist
	}

	return SaveFamilies(dataDir, families)
}

// RemoveMemberFromFamily removes a user from a family. If no members left, deletes the family.
func RemoveMemberFromFamily(dataDir, familyID, username string) error {
	families, err := LoadFamilies(dataDir)
	if err != nil {
		return err
	}

	var newFamilies []Family
	found := false

	for _, f := range families {
		if f.ID == familyID {
			found = true
			var newMembers []string
			for _, m := range f.Members {
				if m != username {
					newMembers = append(newMembers, m)
				}
			}
			f.Members = newMembers
			
			// If family is not empty, keep it
			if len(f.Members) > 0 {
				newFamilies = append(newFamilies, f)
			} else {
				// Family is empty, delete its directory
				_ = os.RemoveAll(GetFamilyDir(dataDir, f.ID))
			}
		} else {
			newFamilies = append(newFamilies, f)
		}
	}

	if !found {
		return os.ErrNotExist
	}

	return SaveFamilies(dataDir, newFamilies)
}
