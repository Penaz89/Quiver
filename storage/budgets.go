package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LoadBudgets reads the category budgets from the data directory.
func LoadBudgets(dataDir string) (map[string]string, error) {
	path := filepath.Join(dataDir, "budgets.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return make(map[string]string), err
	}
	var budgets map[string]string
	if err := json.Unmarshal(data, &budgets); err != nil {
		return make(map[string]string), err
	}
	return budgets, nil
}

// SaveBudgets writes the category budgets to the data directory.
func SaveBudgets(dataDir string, budgets map[string]string) error {
	path := filepath.Join(dataDir, "budgets.json")
	data, err := json.MarshalIndent(budgets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
