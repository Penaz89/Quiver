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
	"menu.habits":   {"en": "HABITS", "it": "ABITUDINI"},
	"menu.journal":  {"en": "JOURNAL", "it": "DIARIO"},
	"menu.tasks":    {"en": "TASKS", "it": "ATTIVITÀ"},
	"menu.vehicles": {"en": "VEHICLES", "it": "VEICOLI"},
	"menu.finances": {"en": "FINANCES", "it": "FINANZE"},
	"menu.weather":  {"en": "WEATHER", "it": "METEO"},
	"menu.vault":    {"en": "VAULT", "it": "VAULT"},
	"menu.settings": {"en": "SETTINGS", "it": "IMPOSTAZIONI"},
	"menu.users":    {"en": "USERS", "it": "UTENTI"},
	"menu.logout":   {"en": "LOGOUT", "it": "LOGOUT"},

	// ── Home view ────────────────────────────────────────────
	"home.welcome":     {"en": "Welcome back, %s!", "it": "Bentornato, %s!"},
	"home.finances":    {"en": "RECURRING EXPENSES SUMMARY", "it": "RESOCONTO SPESE RICORRENTI"},
	"home.annual":      {"en": "Annual:", "it": "Annuale:"},
	"home.monthly":     {"en": "Monthly:", "it": "Mensile:"},
	"home.deadlines":   {"en": "UPCOMING DEADLINES", "it": "PROSSIME SCADENZE"},
	"home.noDeadlines": {"en": "No upcoming deadlines.", "it": "Nessuna scadenza in vista."},
	"home.recentNotes": {"en": "RECENT NOTES", "it": "ULTIME NOTE"},
	"home.noNotes":     {"en": "No journal entries yet.", "it": "Nessuna nota nel diario."},
	"home.recentTasks": {"en": "RECENT TASKS", "it": "ULTIME ATTIVITÀ"},
	"home.noTasks":     {"en": "No tasks found.", "it": "Nessuna attività presente."},
	"home.tagNTC":      {"en": "NTC", "it": "Revisione"},
	"home.tagTax":      {"en": "Tax", "it": "Bollo"},
	"home.tagIns":      {"en": "Ins.", "it": "RCA"},
	"home.tagService":  {"en": "Serv.", "it": "Tagliando"},

	// ── Vehicles section ─────────────────────────────────────
	"vehicles.title":         {"en": "Vehicles", "it": "Veicoli"},
	"vehicles.selectSection": {"en": "Select an option", "it": "Seleziona un'opzione"},
	"vehicles.management":    {"en": "Vehicle Management", "it": "Gestione Veicoli"},
	"vehicles.insurance":     {"en": "Insurance", "it": "Assicurazione"},
	"vehicles.roadTax":       {"en": "Road Tax", "it": "Bollo"},
	"vehicles.ntc":           {"en": "NTC", "it": "Revisione"},
	"vehicles.service":       {"en": "Service", "it": "Tagliando"},
	"vehicles.noVehicles":    {"en": "No vehicles registered yet.", "it": "Nessun veicolo registrato."},
	"vehicles.addFirst":      {"en": "Add vehicles first in Vehicle Management.", "it": "Aggiungi prima i veicoli in Gestione Veicoli."},
	"vehicles.statistics":    {"en": "Statistics", "it": "Statistiche"},
	"vehicles.totalVehicles": {"en": "Total Vehicles:", "it": "Veicoli Totali:"},
	"vehicles.nextExpiry":    {"en": "Next Expiry", "it": "Prossima Scadenza"},
	"vehicles.expiredDays":   {"en": "(Expired %d days ago)", "it": "(Scaduto da %d giorni)"},
	"vehicles.expiresToday":  {"en": "(Expires today)", "it": "(Scade oggi)"},
	"vehicles.expiresIn":     {"en": "(In %d days)", "it": "(Tra %d giorni)"},

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
	"type.monthly":    {"en": "Monthly", "it": "Mensile"},

	// ── CRUD actions ─────────────────────────────────────────
	"action.add":           {"en": "Add", "it": "Aggiungi"},
	"action.edit":          {"en": "Edit", "it": "Modifica"},
	"action.delete":        {"en": "Delete", "it": "Elimina"},
	"action.save":          {"en": "save", "it": "salva"},
	"action.cancel":        {"en": "cancel", "it": "annulla"},
	"action.confirm":       {"en": "confirm", "it": "conferma"},
	"action.back":          {"en": "back", "it": "indietro"},
	"action.locked":        {"en": "locked", "it": "bloccato"},
	"action.selectVeh":     {"en": "Select Vehicle", "it": "Seleziona Veicolo"},
	"action.chooseByPlate": {"en": "Choose a vehicle by license plate", "it": "Scegli un veicolo per targa"},

	// ── Delete confirmation ──────────────────────────────────
	"delete.vehicle":          {"en": "Delete Vehicle", "it": "Elimina Veicolo"},
	"delete.insurance":        {"en": "Delete Insurance", "it": "Elimina Assicurazione"},
	"delete.confirmVehicle":   {"en": "Are you sure you want to delete this vehicle?", "it": "Sei sicuro di voler eliminare questo veicolo?"},
	"delete.confirmInsurance": {"en": "Are you sure you want to delete this record?", "it": "Sei sicuro di voler eliminare questo record?"},

	// ── Insurance ────────────────────────────────────────────
	"insurance.title":     {"en": "Insurance", "it": "Assicurazione"},
	"insurance.noRecords": {"en": "No insurance records yet.", "it": "Nessun record assicurativo."},
	"insurance.add":       {"en": "Add Insurance", "it": "Aggiungi Assicurazione"},
	"insurance.edit":      {"en": "Edit Insurance", "it": "Modifica Assicurazione"},

	// ── Finances view ────────────────────────────────────────
	"finances.title":           {"en": "Finances", "it": "Finanze"},
	"finances.subtitle":        {"en": "Financial log & balance", "it": "Registro finanziario e bilancio"},
	"finances.noEntries":       {"en": "No financial entries yet.", "it": "Nessuna voce finanziaria."},
	"finances.fixedExp":        {"en": "Expenses Summary", "it": "Resoconto Spese"},
	"finances.annualTotal":     {"en": "Annual Total", "it": "Totale Annuale"},
	"finances.monthlyTotal":    {"en": "Monthly Total", "it": "Totale Mensile"},
	"finances.subtotal":        {"en": "Subtotal", "it": "Subtotale"},
	"finances.grandTotal":      {"en": "GRAND TOTAL", "it": "GRAN TOTALE"},
	"finances.housing":         {"en": "Housing", "it": "Casa"},
	"housing.noRecords":        {"en": "No housing expenses yet.", "it": "Nessuna spesa per la casa."},
	"finances.holidays":        {"en": "Holidays", "it": "Vacanze"},
	"holidays.noRecords":       {"en": "No holiday expenses yet.", "it": "Nessuna spesa per vacanze."},
	"finances.subscriptions":   {"en": "Subscriptions", "it": "Abbonamenti"},
	"subscriptions.noRecords":  {"en": "No subscriptions yet.", "it": "Nessun abbonamento."},
	"finances.salaryImpact":    {"en": "SALARY IMPACT", "it": "IMPATTO SULLO STIPENDIO"},
	"finances.projectedAnnual": {"en": "Projected Annual Net", "it": "Netto Annuale Stimato"},
	"finances.fixedAnnual":     {"en": "Total Recurring Expenses", "it": "Totale Spese Ricorrenti"},
	"finances.impactPct":       {"en": "Impact Percentage", "it": "Incidenza Percentuale"},
	"finances.daily":           {"en": "Daily Expenses", "it": "Spese Giornaliere"},
	"daily.noRecords":          {"en": "No daily expenses yet.", "it": "Nessuna spesa giornaliera."},
	"daily.annualTotal":        {"en": "Total Annual:", "it": "Totale Annuo:"},
	"finances.salaries":        {"en": "Salaries", "it": "Stipendi"},
	"salaries.noRecords":       {"en": "No salaries tracked yet.", "it": "Nessuno stipendio tracciato."},
	"salaries.annualSum":       {"en": "Annual Summary", "it": "Resoconto Annuale"},
	"salaries.comparison":      {"en": "Year-over-Year Comparison", "it": "Confronto Anno su Anno"},
	"salaries.difference":      {"en": "Diff vs Prev. Year:", "it": "Diff. su Anno Prec.:"},
	"cat.vehicles":             {"en": "VEHICLES", "it": "VEICOLI"},
	"cat.goals":                {"en": "GOALS", "it": "OBIETTIVI"},
	"col.annual":               {"en": "ANNUAL", "it": "ANNUALE"},
	"col.goal":                 {"en": "GOAL", "it": "GOAL"},
	"col.monthly":              {"en": "MONTHLY", "it": "MENSILE"},
	"col.service":              {"en": "SERVICE", "it": "SERVIZIO"},
	"col.expense":              {"en": "EXPENSE", "it": "SPESA"},
	"col.destination":          {"en": "DESTINATION", "it": "DESTINAZIONE"},
	"col.flight":               {"en": "FLIGHT", "it": "VOLO"},
	"col.accom":                {"en": "ACCOM.", "it": "ALLOGGIO"},
	"col.year":                 {"en": "YEAR", "it": "ANNO"},
	"col.month":                {"en": "MONTH", "it": "MESE"},
	"col.gross":                {"en": "GROSS", "it": "LORDO"},
	"col.net":                  {"en": "NET", "it": "NETTO"},
	"col.deductions":           {"en": "DEDUCTIONS", "it": "TRATTENUTE"},
	"col.taxes":                {"en": "TAXES", "it": "TASSE"},
	"col.amount":               {"en": "AMOUNT", "it": "IMPORTO"},
	"col.date":                 {"en": "DATE", "it": "DATA"},
	"col.category":             {"en": "CATEGORY", "it": "CATEGORIA"},
	"col.description":          {"en": "DESCRIPTION", "it": "DESCRIZIONE"},
	"field.date":               {"en": "Date (DD/MM/YYYY)", "it": "Data (GG/MM/AAAA)"},
	"field.category":           {"en": "Category", "it": "Categoria"},
	"field.description":        {"en": "Description", "it": "Descrizione"},
	"field.amount":             {"en": "Amount", "it": "Importo"},
	"field.service":            {"en": "Service", "it": "Servizio"},
	"field.expense":            {"en": "Expense", "it": "Spesa"},
	"field.destination":        {"en": "Destination", "it": "Destinazione"},
	"field.flightDesc":         {"en": "Flight (Desc)", "it": "Volo (Descrizione)"},
	"field.flightCost":         {"en": "Flight (Cost)", "it": "Volo (Costo)"},
	"field.accomDesc":          {"en": "Accommodation (Desc)", "it": "Alloggio (Descrizione)"},
	"field.accomCost":          {"en": "Accommodation (Cost)", "it": "Alloggio (Costo)"},
	"field.carDesc":            {"en": "Car Rental (Desc)", "it": "Noleggio auto (Descrizione)"},
	"field.carCost":            {"en": "Car Rental (Cost)", "it": "Noleggio auto (Costo)"},
	"field.insDesc":            {"en": "Travel Ins. (Desc)", "it": "Assicurazione Viaggio (Descrizione)"},
	"field.insCost":            {"en": "Travel Ins. (Cost)", "it": "Assicurazione Viaggio (Costo)"},
	"field.year":               {"en": "Year", "it": "Anno"},
	"field.month":              {"en": "Month", "it": "Mese"},
	"field.gross":              {"en": "Gross", "it": "Lordo"},
	"field.net":                {"en": "Net", "it": "Netto"},

	// ── Habits view ──────────────────────────────────────────
	"habits.title":         {"en": "Habits", "it": "Abitudini"},
	"habits.subtitle":      {"en": "Don't break the chain", "it": "Non spezzare la catena"},
	"habits.noItems":       {"en": "No habits tracked yet. Press 'n' to add one.", "it": "Nessuna abitudine tracciata. Premi 'n' per aggiungerne una."},
	"habits.add":           {"en": "Add new habit:", "it": "Aggiungi nuova abitudine:"},
	"habits.name":          {"en": "Habit Name", "it": "Nome abitudine"},
	"habits.addHelp":       {"en": "n: new", "it": "n: nuova"},
	"habits.deleteHelp":    {"en": "d/del: delete", "it": "d/canc: elimina"},
	"habits.toggleHelp":    {"en": "space: toggle today", "it": "spazio: segna oggi"},
	"habits.confirmDelete": {"en": "Delete this habit? (y/N)", "it": "Eliminare questa abitudine? (s/N)"},

	// ── Journal view ─────────────────────────────────────────
	"journal.title":         {"en": "Journal", "it": "Diario"},
	"journal.subtitle":      {"en": "Daily plain text notes", "it": "Note giornaliere in plain text"},
	"journal.exportHelp":    {"en": "e: export Markdown", "it": "e: esporta Markdown"},
	"journal.exported":      {"en": "Exported to:", "it": "Esportato in:"},
	"journal.editHelp":      {"en": "enter: edit", "it": "invio: modifica"},
	"journal.saveHelp":      {"en": "esc: save & back", "it": "esc: salva & indietro"},
	"journal.deleteHelp":    {"en": "d/del: delete", "it": "d/canc: elimina"},
	"journal.confirmDelete": {"en": "Delete this note? (y/N)", "it": "Eliminare questa nota? (s/N)"},
	"journal.navHelp":       {"en": "←/→: change date", "it": "←/→: cambia data"},

	// ── Tasks view ───────────────────────────────────────────
	"tasks.title":         {"en": "Tasks", "it": "Attività"},
	"tasks.subtitle":      {"en": "Getting Things Done", "it": "Getting Things Done"},
	"tasks.todo":          {"en": "TODO", "it": "DA FARE"},
	"tasks.doing":         {"en": "DOING", "it": "IN CORSO"},
	"tasks.done":          {"en": "DONE", "it": "COMPLETATO"},
	"tasks.add":           {"en": "Add new task:", "it": "Aggiungi nuova attività:"},
	"tasks.edit":          {"en": "Edit task:", "it": "Modifica attività:"},
	"tasks.titleLabel":    {"en": "Title", "it": "Titolo"},
	"tasks.projectLabel":  {"en": "Project", "it": "Progetto"},
	"tasks.priorityLabel": {"en": "Priority (H/M/L)", "it": "Priorità (A/M/B)"},
	"tasks.deadlineLabel": {"en": "Deadline (DD/MM/YYYY)", "it": "Scadenza (GG/MM/AAAA)"},
	"tasks.invalidDate":   {"en": "Invalid date format. Use DD/MM/YYYY", "it": "Formato data non valido. Usa GG/MM/AAAA"},
	"tasks.navHelp":       {"en": "h/l: col  j/k: row", "it": "h/l: colonna  j/k: riga"},
	"tasks.addHelp":       {"en": "n: new", "it": "n: nuova"},
	"tasks.moveHelp":      {"en": "J/K: move up/down", "it": "J/K: sposta su/giù"},
	"tasks.statusHelp":    {"en": "H/L: change status", "it": "H/L: cambia stato"},
	"tasks.editHelp":      {"en": "enter: edit", "it": "invio: modifica"},
	"tasks.deleteHelp":    {"en": "d/del: delete", "it": "d/canc: elimina"},

	// ── Vault view ───────────────────────────────────────────
	"vault.title":            {"en": "Vault", "it": "Vault"},
	"vault.subtitle":         {"en": "AES-GCM Encrypted Secrets", "it": "Segreti Cifrati AES-GCM"},
	"vault.enterMasterPwd":   {"en": "Enter Master Password:", "it": "Inserisci Master Password:"},
	"vault.createMasterPwd":  {"en": "Create Master Password (new vault):", "it": "Crea Master Password (nuovo vault):"},
	"vault.add":              {"en": "Add new secret:", "it": "Aggiungi nuovo segreto:"},
	"vault.edit":             {"en": "Edit secret:", "it": "Modifica segreto:"},
	"vault.fieldTitle":       {"en": "Title", "it": "Titolo"},
	"vault.fieldUsername":    {"en": "Username", "it": "Username"},
	"vault.fieldPassword":    {"en": "Password", "it": "Password"},
	"vault.fieldNotes":       {"en": "Notes", "it": "Note"},
	"vault.confirmDelete":    {"en": "Delete this secret? (y/N)", "it": "Eliminare questo segreto? (s/N)"},
	"vault.noSecrets":        {"en": "No secrets stored. Press 'n' to add one.", "it": "Nessun segreto. Premi 'n' per aggiungerne uno."},
	"vault.colTitle":         {"en": "TITLE", "it": "TITOLO"},
	"vault.colUsername":      {"en": "USERNAME", "it": "USERNAME"},
	"vault.help":             {"en": "/: search • n: new • d/del: delete • enter: view/edit • L: lock vault", "it": "/: cerca • n: nuovo • d/canc: elimina • invio: vedi/modifica • L: blocca vault"},
	"vault.helpSearchActive": {"en": "Type to search • Enter/Esc: done", "it": "Digita per cercare • Invio/Esc: fine"},

	// ── Settings view ────────────────────────────────────────
	"settings.title":        {"en": "Settings", "it": "Impostazioni"},
	"settings.subtitle":     {"en": "Application configuration", "it": "Configurazione applicazione"},
	"settings.language":     {"en": "Language", "it": "Lingua"},
	"settings.currentLang":  {"en": "Current:", "it": "Attuale:"},
	"settings.english":      {"en": "English", "it": "Inglese"},
	"settings.italian":      {"en": "Italian", "it": "Italiano"},
	"settings.saved":        {"en": "Settings saved!", "it": "Impostazioni salvate!"},
	"settings.weatherLoc":   {"en": "Weather", "it": "Meteo"},
	"settings.location":     {"en": "Location", "it": "Località"},
	"settings.theme":        {"en": "Theme", "it": "Tema"},
	"settings.currentTheme": {"en": "Current:", "it": "Attuale:"},
	// ── Finances extensions (Goals & Analytics) ──────────────
	"finances.goals":       {"en": "Financial Goals", "it": "Obiettivi"},
	"goals.noRecords":      {"en": "No financial goals set. Press 'n' to add one.", "it": "Nessun obiettivo impostato. Premi 'n' per aggiungerne uno."},
	"goals.expiresIn":      {"en": "Expires in %d days", "it": "Scade tra %d giorni"},
	"goals.expired":        {"en": "Expired", "it": "Scaduto"},
	"goals.savePerMonth":   {"en": "Save: €%.2f/mo", "it": "Da accantonare: €%.2f/mese"},
	"goals.reached":        {"en": "Target reached!", "it": "Obiettivo raggiunto!"},
	"field.goalName":       {"en": "Goal Name", "it": "Nome Obiettivo"},
	"field.goalTarget":     {"en": "Target Amount (€)", "it": "Importo Target (€)"},
	"field.goalCurrent":    {"en": "Current Saved (€)", "it": "Risparmio Attuale (€)"},
	"field.goalDeadline":   {"en": "Deadline", "it": "Scadenza"},
	"delete.confirmGoal":   {"en": "Delete this goal? (y/N)", "it": "Eliminare questo obiettivo? (s/N)"},
	"finances.analytics":   {"en": "Analytics", "it": "Statistiche"},
	"analytics.avgIncome":  {"en": "Avg Monthly Income:", "it": "Entrate Medie Mensili:"},
	"analytics.avgExpense": {"en": "Avg Monthly Expense:", "it": "Uscite Medie Mensili:"},
	"analytics.cashFlow":   {"en": "Monthly Cash Flow:", "it": "Flusso di Cassa Mensile:"},
	"analytics.savingRate": {"en": "Saving Rate:", "it": "Tasso di Risparmio:"},
	"analytics.ratioBar":   {"en": "Income vs Expense Ratio:", "it": "Rapporto Entrate/Uscite:"},

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

	"month.01": {"en": "JAN", "it": "GEN"},
	"month.02": {"en": "FEB", "it": "FEB"},
	"month.03": {"en": "MAR", "it": "MAR"},
	"month.04": {"en": "APR", "it": "APR"},
	"month.05": {"en": "MAY", "it": "MAG"},
	"month.06": {"en": "JUN", "it": "GIU"},
	"month.07": {"en": "JUL", "it": "LUG"},
	"month.08": {"en": "AUG", "it": "AGO"},
	"month.09": {"en": "SEP", "it": "SET"},
	"month.10": {"en": "OCT", "it": "OTT"},
	"month.11": {"en": "NOV", "it": "NOV"},
	"month.12": {"en": "DEC", "it": "DIC"},
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
