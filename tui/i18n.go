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

package tui

// ─── Translation keys ───────────────────────────────────────────────

// All translatable UI strings are stored here, keyed by identifier.
// Add new keys as needed when extending the interface.

var translations = map[string]map[string]string{
	// ── Main menu ────────────────────────────────────────────
	"menu.home":     {"en": "HOME", "it": "HOME"},
	"menu.vehicles": {"en": "VEHICLES", "it": "VEICOLI"},
	"menu.finances": {"en": "FINANCES", "it": "FINANZE"},
	"menu.settings": {"en": "SETTINGS", "it": "IMPOSTAZIONI"},

	// ── Home view ────────────────────────────────────────────
	"home.welcome":      {"en": "Welcome back, %s!", "it": "Bentornato, %s!"},
	"home.terminal":     {"en": "Terminal:", "it": "Terminale:"},
	"home.window":       {"en": "Window:", "it": "Finestra:"},
	"home.background":   {"en": "Background:", "it": "Sfondo:"},
	"home.colorProfile": {"en": "Color Profile:", "it": "Profilo Colore:"},
	"home.dataDir":      {"en": "Data Directory:", "it": "Cartella Dati:"},

	// ── Vehicles section ─────────────────────────────────────
	"vehicles.title":          {"en": "Vehicles", "it": "Veicoli"},
	"vehicles.selectSection":  {"en": "Select an option", "it": "Seleziona un'opzione"},
	"vehicles.management":     {"en": "Vehicle Management", "it": "Gestione Veicoli"},
	"vehicles.insurance":      {"en": "Insurance", "it": "Assicurazione"},
	"vehicles.roadTax":        {"en": "Road Tax", "it": "Bollo"},
	"vehicles.ntc":            {"en": "NTC", "it": "Revisione"},
	"vehicles.noVehicles":     {"en": "No vehicles registered yet.", "it": "Nessun veicolo registrato."},
	"vehicles.addFirst":       {"en": "Add vehicles first in Vehicle Management.", "it": "Aggiungi prima i veicoli in Gestione Veicoli."},
	"vehicles.statistics":     {"en": "Statistics", "it": "Statistiche"},
	"vehicles.totalVehicles":  {"en": "Total Vehicles:", "it": "Veicoli Totali:"},
	"vehicles.nextExpiry":     {"en": "Next Expiry", "it": "Prossima Scadenza"},
	"vehicles.expiredDays":    {"en": "(Expired %d days ago)", "it": "(Scaduto da %d giorni)"},
	"vehicles.expiresToday":   {"en": "(Expires today)", "it": "(Scade oggi)"},
	"vehicles.expiresIn":      {"en": "(In %d days)", "it": "(Tra %d giorni)"},

	// ── Vehicle form ─────────────────────────────────────────
	"field.brand":        {"en": "Brand", "it": "Marca"},
	"field.model":        {"en": "Model", "it": "Modello"},
	"field.licensePlate": {"en": "License Plate", "it": "Targa"},
	"field.owner":        {"en": "Owner", "it": "Proprietario"},
	"field.totalCost":    {"en": "Total Cost", "it": "Costo Totale"},
	"field.expireDate":   {"en": "Expire Date", "it": "Data Scadenza"},
	"field.insType":      {"en": "Type", "it": "Tipologia"},

	// ── Table column headers ─────────────────────────────────
	"col.num":     {"en": "#", "it": "#"},
	"col.brand":   {"en": "BRAND", "it": "MARCA"},
	"col.model":   {"en": "MODEL", "it": "MODELLO"},
	"col.plate":   {"en": "PLATE", "it": "TARGA"},
	"col.owner":   {"en": "OWNER", "it": "PROPRIETARIO"},
	"col.cost":    {"en": "COST", "it": "COSTO"},
	"col.expires": {"en": "EXPIRES", "it": "SCADENZA"},
	"col.type":    {"en": "TYPE", "it": "TIPOLOGIA"},

	// ── Types ────────────────────────────────────────────────
	"type.semiannual": {"en": "Semiannual", "it": "Semestrale"},
	"type.annual":     {"en": "Annual", "it": "Annuale"},

	// ── CRUD actions ─────────────────────────────────────────
	"action.add":       {"en": "Add", "it": "Aggiungi"},
	"action.edit":      {"en": "Edit", "it": "Modifica"},
	"action.delete":    {"en": "Delete", "it": "Elimina"},
	"action.save":      {"en": "save", "it": "salva"},
	"action.cancel":    {"en": "cancel", "it": "annulla"},
	"action.confirm":   {"en": "confirm", "it": "conferma"},
	"action.back":      {"en": "back", "it": "indietro"},
	"action.locked":    {"en": "locked", "it": "bloccato"},
	"action.selectVeh": {"en": "Select Vehicle", "it": "Seleziona Veicolo"},
	"action.chooseByPlate": {"en": "Choose a vehicle by license plate", "it": "Scegli un veicolo per targa"},

	// ── Delete confirmation ──────────────────────────────────
	"delete.vehicle":   {"en": "Delete Vehicle", "it": "Elimina Veicolo"},
	"delete.insurance": {"en": "Delete Insurance", "it": "Elimina Assicurazione"},
	"delete.confirmVehicle":   {"en": "Are you sure you want to delete this vehicle?", "it": "Sei sicuro di voler eliminare questo veicolo?"},
	"delete.confirmInsurance": {"en": "Are you sure you want to delete this record?", "it": "Sei sicuro di voler eliminare questo record?"},

	// ── Insurance ────────────────────────────────────────────
	"insurance.title":    {"en": "Insurance", "it": "Assicurazione"},
	"insurance.noRecords": {"en": "No insurance records yet.", "it": "Nessun record assicurativo."},
	"insurance.add":      {"en": "Add Insurance", "it": "Aggiungi Assicurazione"},
	"insurance.edit":     {"en": "Edit Insurance", "it": "Modifica Assicurazione"},

	// ── Finances view ────────────────────────────────────────
	"finances.title":       {"en": "Finances", "it": "Finanze"},
	"finances.subtitle":    {"en": "Financial log & balance", "it": "Registro finanziario e bilancio"},
	"finances.noEntries":   {"en": "No financial entries yet.", "it": "Nessuna voce finanziaria."},
	"finances.fixedExp":    {"en": "Fixed Expenses", "it": "Spese Fisse"},
	"finances.annualTotal": {"en": "Annual Total", "it": "Totale Annuale"},
	"finances.monthlyTotal":{"en": "Monthly Total", "it": "Totale Mensile"},
	"finances.subtotal":    {"en": "Subtotal", "it": "Subtotale"},
	"finances.grandTotal":  {"en": "GRAND TOTAL", "it": "GRAN TOTALE"},
	"cat.vehicles":         {"en": "VEHICLES", "it": "VEICOLI"},
	"col.annual":           {"en": "ANNUAL", "it": "ANNUALE"},
	"col.monthly":          {"en": "MONTHLY", "it": "MENSILE"},

	// ── Settings view ────────────────────────────────────────
	"settings.title":       {"en": "Settings", "it": "Impostazioni"},
	"settings.subtitle":    {"en": "Application configuration", "it": "Configurazione applicazione"},
	"settings.language":    {"en": "Language", "it": "Lingua"},
	"settings.currentLang": {"en": "Current:", "it": "Attuale:"},
	"settings.english":     {"en": "English", "it": "Inglese"},
	"settings.italian":     {"en": "Italian", "it": "Italiano"},
	"settings.saved":       {"en": "Settings saved!", "it": "Impostazioni salvate!"},
	"settings.weatherLoc":  {"en": "Weather", "it": "Meteo"},
	"settings.location":    {"en": "Location", "it": "Località"},

	// ── Help bar ─────────────────────────────────────────────
	"help.navigate":       {"en": "navigate", "it": "naviga"},
	"help.enter":          {"en": "enter", "it": "entra"},
	"help.goBack":         {"en": "back", "it": "indietro"},
	"help.focusContent":   {"en": "focus content", "it": "focus contenuto"},
	"help.quit":           {"en": "quit", "it": "esci"},
	"help.menu":           {"en": "menu", "it": "menu"},
	"help.contentFocused": {"en": "content focused", "it": "contenuto attivo"},
	"help.switchField":    {"en": "switch field", "it": "cambia campo"},
	"help.select":         {"en": "select", "it": "seleziona"},
}

// t returns the translation for the given key and language.
// Falls back to English if the language or key is not found.
func t(lang, key string) string {
	if entry, ok := translations[key]; ok {
		if val, ok := entry[lang]; ok {
			return val
		}
		if val, ok := entry["en"]; ok {
			return val
		}
	}
	return key // fallback: return the key itself
}
