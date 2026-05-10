package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Installment represents a recurring expense or fixed-term installment.
type Installment struct {
	Name       string    `json:"name"`
	Amount     string    `json:"amount"`     // Amount per period
	TotalCount int       `json:"totalCount"` // 0 for indefinite
	PaidCount  int       `json:"paidCount"`  // How many times it has been paid
	Frequency  string    `json:"frequency"`  // "monthly", "bimonthly", "quarterly", "semiannual", "annual"
	StartDate  time.Time `json:"startDate"`
	Account    string    `json:"account,omitempty"`
	Author     string    `json:"author"`
}

func LoadInstallments(dataDir string) ([]Installment, error) {
	path := filepath.Join(dataDir, "installments.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Installment{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []Installment{}, err
	}
	var insts []Installment
	if err := json.Unmarshal(data, &insts); err != nil {
		return []Installment{}, err
	}
	return insts, nil
}

func SaveInstallments(dataDir string, insts []Installment) error {
	path := filepath.Join(dataDir, "installments.json")
	data, err := json.MarshalIndent(insts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
