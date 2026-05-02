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
)

// Housing represents a recurring housing expense.
type Housing struct {
	Expense string `json:"expense"`
	Cost    string `json:"cost"`
	Type    string `json:"type"` // "type.monthly" or "type.annual"
}

func housingPath(dataDir string) string {
	return filepath.Join(dataDir, "housing.json")
}

// LoadHousing reads housing expense records from the data directory.
func LoadHousing(dataDir string) ([]Housing, error) {
	path := housingPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Housing{}, nil
		}
		return nil, err
	}

	var exps []Housing
	if err := json.Unmarshal(data, &exps); err != nil {
		return nil, err
	}
	return exps, nil
}

// SaveHousing writes housing expense records to the data directory.
func SaveHousing(dataDir string, exps []Housing) error {
	path := housingPath(dataDir)
	data, err := json.MarshalIndent(exps, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
