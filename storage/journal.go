// Quiver - An SSH TUI Application
// Copyright (C) 2026  penaz
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Journal represents a collection of daily entries.
type Journal struct {
	Entries map[string]string `json:"entries"` // format: "YYYY-MM-DD" -> text
}

func journalPath(dataDir string) string {
	return filepath.Join(dataDir, "journal.json")
}

// LoadJournal reads the journal from the data directory.
func LoadJournal(dataDir string) (Journal, error) {
	path := journalPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Journal{Entries: make(map[string]string)}, nil
		}
		return Journal{}, err
	}
	var j Journal
	if err := json.Unmarshal(data, &j); err != nil {
		return Journal{}, err
	}
	if j.Entries == nil {
		j.Entries = make(map[string]string)
	}
	return j, nil
}

// SaveJournal writes the journal to the data directory.
func SaveJournal(dataDir string, j Journal) error {
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(journalPath(dataDir), data, 0644)
}

// ExportJournalMarkdown exports the journal to a Markdown file.
func ExportJournalMarkdown(dataDir string, j Journal) (string, error) {
	exportPath := filepath.Join(dataDir, "journal_export.md")
	
	// Sort dates
	var dates []string
	for date := range j.Entries {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	
	f, err := os.Create(exportPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	
	fmt.Fprintln(f, "# My Journal")
	fmt.Fprintln(f, "")
	
	for _, date := range dates {
		entry := j.Entries[date]
		if entry == "" {
			continue
		}
		fmt.Fprintf(f, "## %s\n\n", date)
		fmt.Fprintln(f, entry)
		fmt.Fprintln(f, "")
	}
	
	return exportPath, nil
}
