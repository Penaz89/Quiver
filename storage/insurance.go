// Quiver - An SSH TUI Application
// Copyright (C) 2026  penaz
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Insurance represents an insurance policy linked to a vehicle by license plate.
type Insurance struct {
	LicensePlate string    `json:"license_plate"`
	TotalCost    string    `json:"total_cost"`
	ExpireDate   time.Time `json:"expire_date"`
}

func insurancePath(dataDir string) string {
	return filepath.Join(dataDir, "insurance.json")
}

// LoadInsurance reads insurance records from the data directory.
func LoadInsurance(dataDir string) ([]Insurance, error) {
	path := insurancePath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Insurance{}, nil
		}
		return nil, err
	}
	var records []Insurance
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// SaveInsurance writes insurance records to the data directory.
func SaveInsurance(dataDir string, records []Insurance) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(insurancePath(dataDir), data, 0644)
}
