package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Account represents a financial account (e.g. Bank, Cash, Crypto).
type Account struct {
	Name    string `json:"name"`
	Balance string `json:"balance"`
	Type    string `json:"type"`
	Author  string `json:"author"`
}

func LoadAccounts(dataDir string) ([]Account, error) {
	path := filepath.Join(dataDir, "accounts.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Account{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []Account{}, err
	}
	var accounts []Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return []Account{}, err
	}
	return accounts, nil
}

func SaveAccounts(dataDir string, accounts []Account) error {
	path := filepath.Join(dataDir, "accounts.json")
	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
