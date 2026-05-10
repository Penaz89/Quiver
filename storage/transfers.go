package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Transfer struct {
	Date        time.Time `json:"date"`
	FromAccount string    `json:"fromAccount"`
	ToAccount   string    `json:"toAccount"`
	Amount      string    `json:"amount"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Frequency   string    `json:"frequency,omitempty"` // "none", "monthly"
	NextDate    time.Time `json:"nextDate,omitempty"`
}

func LoadTransfers(dataDir string) ([]Transfer, error) {
	path := filepath.Join(dataDir, "transfers.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Transfer{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t []Transfer
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return t, nil
}

func SaveTransfers(dataDir string, transfers []Transfer) error {
	path := filepath.Join(dataDir, "transfers.json")
	b, err := json.MarshalIndent(transfers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
