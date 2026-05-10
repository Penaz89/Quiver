package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LoadCategories reads the list of daily expense categories from the data directory.
func LoadCategories(dataDir string) ([]string, error) {
	path := filepath.Join(dataDir, "categories.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Provide some defaults if not present
		defaultCats := []string{"Spesa", "Bar/Ristorante", "Shopping", "Salute", "Trasporti", "Varie"}
		_ = SaveCategories(dataDir, defaultCats)
		return defaultCats, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{}, err
	}
	var cats []string
	if err := json.Unmarshal(data, &cats); err != nil {
		return []string{}, err
	}
	return cats, nil
}

// SaveCategories writes the list of daily expense categories to the data directory.
func SaveCategories(dataDir string, cats []string) error {
	path := filepath.Join(dataDir, "categories.json")
	data, err := json.MarshalIndent(cats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
