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

// Holiday represents a holiday/vacation budget or expense.
type Holiday struct {
	Destination string `json:"destination"`
	FlightDesc  string `json:"flight_desc"`
	FlightCost  string `json:"flight_cost"`
	AccomDesc   string `json:"accom_desc"`
	AccomCost   string `json:"accom_cost"`
	CarDesc     string `json:"car_desc"`
	CarCost     string `json:"car_cost"`
	InsDesc     string `json:"ins_desc"`
	InsCost     string `json:"ins_cost"`
}

func holidaysPath(dataDir string) string {
	return filepath.Join(dataDir, "holidays.json")
}

// LoadHolidays reads holiday records from the data directory.
func LoadHolidays(dataDir string) ([]Holiday, error) {
	path := holidaysPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Holiday{}, nil
		}
		return nil, err
	}

	var exps []Holiday
	if err := json.Unmarshal(data, &exps); err != nil {
		return nil, err
	}
	return exps, nil
}

// SaveHolidays writes holiday records to the data directory.
func SaveHolidays(dataDir string, exps []Holiday) error {
	path := holidaysPath(dataDir)
	data, err := json.MarshalIndent(exps, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
