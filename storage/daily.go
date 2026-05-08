package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type DailyExpense struct {
	Date        time.Time `json:"date"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Amount      string    `json:"amount"`
}

func LoadDailyExpenses(dataDir string) ([]DailyExpense, error) {
	path := filepath.Join(dataDir, "daily.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []DailyExpense{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []DailyExpense{}, err
	}
	var expenses []DailyExpense
	if err := json.Unmarshal(data, &expenses); err != nil {
		return []DailyExpense{}, err
	}
	return expenses, nil
}

func SaveDailyExpenses(dataDir string, expenses []DailyExpense) error {
	path := filepath.Join(dataDir, "daily.json")
	data, err := json.MarshalIndent(expenses, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
