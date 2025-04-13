package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed static
var content embed.FS

// Version information - set during build
var (
	Version    = "dev"
	BuildTime  = ""
	CommitHash = ""
)

const (
	dbFile = "condo.db"
	port   = "8080"
)

// Models
type Resident struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	Contact   string    `json:"contact"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Payment struct {
	ID           int       `json:"id"`
	ResidentID   int       `json:"resident_id"`
	ResidentName string    `json:"residentName,omitempty"`
	Amount       float64   `json:"amount"`
	Description  string    `json:"description"`
	PaymentDate  string    `json:"payment_date"`
	CreatedAt    time.Time `json:"created_at"`
}

type Expense struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	ExpenseDate string    `json:"expense_date"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// ExportData represents the entire database structure for export/import
type ExportData struct {
	Residents  []Resident `json:"residents"`
	Payments   []Payment  `json:"payments"`
	Expenses   []Expense  `json:"expenses"`
	ExportDate string     `json:"export_date"`
}

func main() {
	// Parse command-line flags
	loadSampleData := flag.Bool("sample", false, "Load sample data into the database")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("Condo Manager %s\n", Version)
		if BuildTime != "" {
			fmt.Printf("Build Time: %s\n", BuildTime)
		}
		if CommitHash != "" {
			fmt.Printf("Commit: %s\n", CommitHash)
		}
		return
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Load sample data if requested
	if *loadSampleData {
		err := insertSampleData(db)
		if err != nil {
			log.Printf("Warning: Failed to load sample data: %v", err)
		} else {
			log.Println("Sample data loaded successfully")
		}
	}

	// Initialize router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	// Residents API endpoints
	api.HandleFunc("/residents", getResidents(db)).Methods("GET")
	api.HandleFunc("/residents", createResident(db)).Methods("POST")
	api.HandleFunc("/residents/{id:[0-9]+}", getResident(db)).Methods("GET")
	api.HandleFunc("/residents/{id:[0-9]+}", updateResident(db)).Methods("PUT")
	api.HandleFunc("/residents/{id:[0-9]+}", deleteResident(db)).Methods("DELETE")

	// Payments API endpoints
	api.HandleFunc("/payments", getPayments(db)).Methods("GET")
	api.HandleFunc("/payments", createPayment(db)).Methods("POST")
	api.HandleFunc("/payments/{id:[0-9]+}", getPayment(db)).Methods("GET")
	api.HandleFunc("/payments/{id:[0-9]+}", updatePayment(db)).Methods("PUT")
	api.HandleFunc("/payments/{id:[0-9]+}", deletePayment(db)).Methods("DELETE")

	// Expenses API endpoints
	api.HandleFunc("/expenses", getExpenses(db)).Methods("GET")
	api.HandleFunc("/expenses", createExpense(db)).Methods("POST")
	api.HandleFunc("/expenses/{id:[0-9]+}", getExpense(db)).Methods("GET")
	api.HandleFunc("/expenses/{id:[0-9]+}", updateExpense(db)).Methods("PUT")
	api.HandleFunc("/expenses/{id:[0-9]+}", deleteExpense(db)).Methods("DELETE")

	// Export and Import API endpoints
	api.HandleFunc("/export", exportDatabase(db)).Methods("GET")
	api.HandleFunc("/import", importDatabase(db)).Methods("POST")

	// Search API endpoints
	api.HandleFunc("/search/residents", searchResidents(db)).Methods("GET")
	api.HandleFunc("/search/payments", searchPayments(db)).Methods("GET")
	api.HandleFunc("/search/expenses", searchExpenses(db)).Methods("GET")

	// Reports Export endpoints
	api.HandleFunc("/reports/payments/export", exportPaymentsReport(db)).Methods("GET")
	api.HandleFunc("/reports/expenses/export", exportExpensesReport(db)).Methods("GET")

	// Serve static files
	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(content)))

	// Serve index page
	r.PathPrefix("/").HandlerFunc(serveIndex)

	// Start server
	fmt.Printf("Server is running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initDB() (*sql.DB, error) {
	// Create database directory if it doesn't exist
	dbDir := filepath.Dir(dbFile)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %v", err)
		}
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create tables if they don't exist
	err = createTables(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	// Create residents table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS residents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			unit TEXT NOT NULL,
			contact TEXT,
			email TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create payments table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			resident_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			description TEXT,
			payment_date DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (resident_id) REFERENCES residents (id)
		)
	`)
	if err != nil {
		return err
	}

	// Create expenses table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount REAL NOT NULL,
			description TEXT,
			expense_date DATE NOT NULL,
			category TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	data, err := content.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Could not load page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Validation function for Resident data
func validateResident(r Resident) error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Unit == "" {
		return fmt.Errorf("unit is required")
	}
	if r.Email != "" {
		// Simple email validation
		if !strings.Contains(r.Email, "@") || !strings.Contains(r.Email, ".") {
			return fmt.Errorf("invalid email format")
		}
	}
	return nil
}

// Validation function for Payment data
func validatePayment(p Payment) error {
	if p.ResidentID <= 0 {
		return fmt.Errorf("resident is required")
	}
	if p.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if p.PaymentDate == "" {
		return fmt.Errorf("payment date is required")
	}
	// Validate date format
	_, err := time.Parse("2006-01-02", p.PaymentDate)
	if err != nil {
		return fmt.Errorf("invalid date format, must be YYYY-MM-DD")
	}
	return nil
}

// Validation function for Expense data
func validateExpense(e Expense) error {
	if e.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if e.Description == "" {
		return fmt.Errorf("description is required")
	}
	if e.ExpenseDate == "" {
		return fmt.Errorf("expense date is required")
	}
	// Validate date format
	_, err := time.Parse("2006-01-02", e.ExpenseDate)
	if err != nil {
		return fmt.Errorf("invalid date format, must be YYYY-MM-DD")
	}
	return nil
}

// Handlers for resident endpoints
func getResidents(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, unit, contact, email, created_at, updated_at FROM residents ORDER BY name")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		residents := []Resident{}
		for rows.Next() {
			var resident Resident
			if err := rows.Scan(&resident.ID, &resident.Name, &resident.Unit, &resident.Contact, &resident.Email, &resident.CreatedAt, &resident.UpdatedAt); err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			residents = append(residents, resident)
		}

		respondWithJSON(w, http.StatusOK, residents)
	}
}

func createResident(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resident Resident
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&resident); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate resident data
		if err := validateResident(resident); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		stmt, err := db.Prepare("INSERT INTO residents(name, unit, contact, email) VALUES(?, ?, ?, ?)")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(resident.Name, resident.Unit, resident.Contact, resident.Email)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		resident.ID = int(id)
		respondWithJSON(w, http.StatusCreated, resident)
	}
}

func getResident(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid resident ID")
			return
		}

		var resident Resident
		err = db.QueryRow("SELECT id, name, unit, contact, email, created_at, updated_at FROM residents WHERE id = ?", id).
			Scan(&resident.ID, &resident.Name, &resident.Unit, &resident.Contact, &resident.Email, &resident.CreatedAt, &resident.UpdatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusNotFound, "Resident not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, resident)
	}
}

func updateResident(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid resident ID")
			return
		}

		var resident Resident
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&resident); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate resident data
		if err := validateResident(resident); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		stmt, err := db.Prepare("UPDATE residents SET name = ?, unit = ?, contact = ?, email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(resident.Name, resident.Unit, resident.Contact, resident.Email, id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		resident.ID = id
		respondWithJSON(w, http.StatusOK, resident)
	}
}

func deleteResident(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid resident ID")
			return
		}

		stmt, err := db.Prepare("DELETE FROM residents WHERE id = ?")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	}
}

// Handlers for payment endpoints
func getPayments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT p.id, p.resident_id, r.name, p.amount, p.description, p.payment_date, p.created_at 
			FROM payments p
			JOIN residents r ON p.resident_id = r.id
			ORDER BY p.payment_date DESC
		`)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		payments := []Payment{}
		for rows.Next() {
			var payment Payment
			if err := rows.Scan(&payment.ID, &payment.ResidentID, &payment.ResidentName, &payment.Amount, &payment.Description, &payment.PaymentDate, &payment.CreatedAt); err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			payments = append(payments, payment)
		}

		respondWithJSON(w, http.StatusOK, payments)
	}
}

func createPayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payment Payment
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payment); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate payment data
		if err := validatePayment(payment); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		stmt, err := db.Prepare("INSERT INTO payments(resident_id, amount, description, payment_date) VALUES(?, ?, ?, ?)")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(payment.ResidentID, payment.Amount, payment.Description, payment.PaymentDate)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		payment.ID = int(id)
		respondWithJSON(w, http.StatusCreated, payment)
	}
}

func getPayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid payment ID")
			return
		}

		var payment Payment
		err = db.QueryRow(`
			SELECT p.id, p.resident_id, r.name, p.amount, p.description, p.payment_date, p.created_at 
			FROM payments p
			JOIN residents r ON p.resident_id = r.id
			WHERE p.id = ?
		`, id).Scan(&payment.ID, &payment.ResidentID, &payment.ResidentName, &payment.Amount, &payment.Description, &payment.PaymentDate, &payment.CreatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusNotFound, "Payment not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, payment)
	}
}

func updatePayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid payment ID")
			return
		}

		var payment Payment
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payment); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate payment data
		if err := validatePayment(payment); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		stmt, err := db.Prepare("UPDATE payments SET resident_id = ?, amount = ?, description = ?, payment_date = ? WHERE id = ?")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(payment.ResidentID, payment.Amount, payment.Description, payment.PaymentDate, id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		payment.ID = id
		respondWithJSON(w, http.StatusOK, payment)
	}
}

func deletePayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid payment ID")
			return
		}

		stmt, err := db.Prepare("DELETE FROM payments WHERE id = ?")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	}
}

// Handlers for expense endpoints
func getExpenses(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, amount, description, expense_date, category, created_at FROM expenses ORDER BY expense_date DESC")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		expenses := []Expense{}
		for rows.Next() {
			var expense Expense
			if err := rows.Scan(&expense.ID, &expense.Amount, &expense.Description, &expense.ExpenseDate, &expense.Category, &expense.CreatedAt); err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			expenses = append(expenses, expense)
		}

		respondWithJSON(w, http.StatusOK, expenses)
	}
}

func createExpense(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var expense Expense
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&expense); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate expense data
		if err := validateExpense(expense); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		stmt, err := db.Prepare("INSERT INTO expenses(amount, description, expense_date, category) VALUES(?, ?, ?, ?)")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(expense.Amount, expense.Description, expense.ExpenseDate, expense.Category)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		expense.ID = int(id)
		respondWithJSON(w, http.StatusCreated, expense)
	}
}

func getExpense(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
			return
		}

		var expense Expense
		err = db.QueryRow("SELECT id, amount, description, expense_date, category, created_at FROM expenses WHERE id = ?", id).
			Scan(&expense.ID, &expense.Amount, &expense.Description, &expense.ExpenseDate, &expense.Category, &expense.CreatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusNotFound, "Expense not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, expense)
	}
}

func updateExpense(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
			return
		}

		var expense Expense
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&expense); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate expense data
		if err := validateExpense(expense); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		stmt, err := db.Prepare("UPDATE expenses SET amount = ?, description = ?, expense_date = ?, category = ? WHERE id = ?")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(expense.Amount, expense.Description, expense.ExpenseDate, expense.Category, id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		expense.ID = id
		respondWithJSON(w, http.StatusOK, expense)
	}
}

func deleteExpense(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
			return
		}

		stmt, err := db.Prepare("DELETE FROM expenses WHERE id = ?")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	}
}

// Export database as JSON
func exportDatabase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exportData := ExportData{
			ExportDate: time.Now().Format(time.RFC3339),
		}

		// Get all residents
		residents, err := getAllResidents(db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error exporting residents: %v", err))
			return
		}
		exportData.Residents = residents

		// Get all payments
		payments, err := getAllPayments(db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error exporting payments: %v", err))
			return
		}
		exportData.Payments = payments

		// Get all expenses
		expenses, err := getAllExpenses(db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error exporting expenses: %v", err))
			return
		}
		exportData.Expenses = expenses

		// Set header for file download
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=condo_export_%s.json",
			time.Now().Format("2006-01-02")))

		// Write JSON response
		if err := json.NewEncoder(w).Encode(exportData); err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error encoding export data: %v", err))
			return
		}
	}
}

// Import database from JSON
func importDatabase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
			respondWithError(w, http.StatusBadRequest, "Unable to parse form")
			return
		}

		// Get file from form
		file, _, err := r.FormFile("importFile")
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Error retrieving import file")
			return
		}
		defer file.Close()

		// Read file content
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error reading import file")
			return
		}

		// Parse JSON data
		var importData ExportData
		if err := json.Unmarshal(fileBytes, &importData); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid import file format")
			return
		}

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to begin transaction")
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
				return
			}
		}()

		// Clear existing data
		if _, err = tx.Exec("DELETE FROM payments"); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to clear existing payments")
			return
		}
		if _, err = tx.Exec("DELETE FROM expenses"); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to clear existing expenses")
			return
		}
		if _, err = tx.Exec("DELETE FROM residents"); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to clear existing residents")
			return
		}

		// Insert residents
		stmt, err := tx.Prepare("INSERT INTO residents(id, name, unit, contact, email) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to prepare resident statement")
			return
		}
		defer stmt.Close()

		for _, resident := range importData.Residents {
			_, err := stmt.Exec(resident.ID, resident.Name, resident.Unit, resident.Contact, resident.Email)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to import resident: %v", err))
				return
			}
		}

		// Insert payments
		stmt, err = tx.Prepare("INSERT INTO payments(id, resident_id, amount, description, payment_date) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to prepare payment statement")
			return
		}
		defer stmt.Close()

		for _, payment := range importData.Payments {
			_, err := stmt.Exec(payment.ID, payment.ResidentID, payment.Amount, payment.Description, payment.PaymentDate)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to import payment: %v", err))
				return
			}
		}

		// Insert expenses
		stmt, err = tx.Prepare("INSERT INTO expenses(id, amount, description, expense_date, category) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to prepare expense statement")
			return
		}
		defer stmt.Close()

		for _, expense := range importData.Expenses {
			_, err := stmt.Exec(expense.ID, expense.Amount, expense.Description, expense.ExpenseDate, expense.Category)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to import expense: %v", err))
				return
			}
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{
			"message":            "Database import successful",
			"imported_residents": strconv.Itoa(len(importData.Residents)),
			"imported_payments":  strconv.Itoa(len(importData.Payments)),
			"imported_expenses":  strconv.Itoa(len(importData.Expenses)),
		})
	}
}

// Helper function to get all residents
func getAllResidents(db *sql.DB) ([]Resident, error) {
	rows, err := db.Query("SELECT id, name, unit, contact, email, created_at, updated_at FROM residents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	residents := []Resident{}
	for rows.Next() {
		var resident Resident
		if err := rows.Scan(&resident.ID, &resident.Name, &resident.Unit, &resident.Contact, &resident.Email, &resident.CreatedAt, &resident.UpdatedAt); err != nil {
			return nil, err
		}
		residents = append(residents, resident)
	}

	return residents, nil
}

// Helper function to get all payments
func getAllPayments(db *sql.DB) ([]Payment, error) {
	rows, err := db.Query("SELECT id, resident_id, amount, description, payment_date, created_at FROM payments")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		var payment Payment
		if err := rows.Scan(&payment.ID, &payment.ResidentID, &payment.Amount, &payment.Description, &payment.PaymentDate, &payment.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

// Helper function to get all expenses
func getAllExpenses(db *sql.DB) ([]Expense, error) {
	rows, err := db.Query("SELECT id, amount, description, expense_date, category, created_at FROM expenses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := []Expense{}
	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.ID, &expense.Amount, &expense.Description, &expense.ExpenseDate, &expense.Category, &expense.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

// Search for residents
func searchResidents(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			respondWithError(w, http.StatusBadRequest, "Search query is required")
			return
		}

		// SQL query with LIKE for matching name, unit, or email
		sqlQuery := `
			SELECT id, name, unit, contact, email, created_at, updated_at 
			FROM residents 
			WHERE name LIKE ? OR unit LIKE ? OR email LIKE ? OR contact LIKE ?
			ORDER BY name
		`
		searchPattern := "%" + query + "%"

		rows, err := db.Query(sqlQuery, searchPattern, searchPattern, searchPattern, searchPattern)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		residents := []Resident{}
		for rows.Next() {
			var resident Resident
			if err := rows.Scan(&resident.ID, &resident.Name, &resident.Unit, &resident.Contact, &resident.Email, &resident.CreatedAt, &resident.UpdatedAt); err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			residents = append(residents, resident)
		}

		respondWithJSON(w, http.StatusOK, residents)
	}
}

// Search for payments
func searchPayments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		residentId := r.URL.Query().Get("resident_id")
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		// Build WHERE clause dynamically
		whereClause := ""
		args := []interface{}{}

		if query != "" {
			whereClause += "p.description LIKE ? OR r.name LIKE ?"
			searchPattern := "%" + query + "%"
			args = append(args, searchPattern, searchPattern)
		}

		if residentId != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "p.resident_id = ?"
			args = append(args, residentId)
		}

		if startDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "p.payment_date >= ?"
			args = append(args, startDate)
		}

		if endDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "p.payment_date <= ?"
			args = append(args, endDate)
		}

		// Build full SQL query
		sqlQuery := `
			SELECT p.id, p.resident_id, r.name, p.amount, p.description, p.payment_date, p.created_at 
			FROM payments p
			JOIN residents r ON p.resident_id = r.id
		`

		if whereClause != "" {
			sqlQuery += " WHERE " + whereClause
		}

		sqlQuery += " ORDER BY p.payment_date DESC"

		rows, err := db.Query(sqlQuery, args...)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		payments := []Payment{}
		for rows.Next() {
			var payment Payment
			if err := rows.Scan(&payment.ID, &payment.ResidentID, &payment.ResidentName, &payment.Amount, &payment.Description, &payment.PaymentDate, &payment.CreatedAt); err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			payments = append(payments, payment)
		}

		respondWithJSON(w, http.StatusOK, payments)
	}
}

// Search for expenses
func searchExpenses(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		category := r.URL.Query().Get("category")
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		// Build WHERE clause dynamically
		whereClause := ""
		args := []interface{}{}

		if query != "" {
			whereClause += "description LIKE ?"
			searchPattern := "%" + query + "%"
			args = append(args, searchPattern)
		}

		if category != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "category = ?"
			args = append(args, category)
		}

		if startDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "expense_date >= ?"
			args = append(args, startDate)
		}

		if endDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "expense_date <= ?"
			args = append(args, endDate)
		}

		// Build full SQL query
		sqlQuery := "SELECT id, amount, description, expense_date, category, created_at FROM expenses"

		if whereClause != "" {
			sqlQuery += " WHERE " + whereClause
		}

		sqlQuery += " ORDER BY expense_date DESC"

		rows, err := db.Query(sqlQuery, args...)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		expenses := []Expense{}
		for rows.Next() {
			var expense Expense
			if err := rows.Scan(&expense.ID, &expense.Amount, &expense.Description, &expense.ExpenseDate, &expense.Category, &expense.CreatedAt); err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			expenses = append(expenses, expense)
		}

		respondWithJSON(w, http.StatusOK, expenses)
	}
}

// Export payments report as CSV
func exportPaymentsReport(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters for filtering
		residentId := r.URL.Query().Get("resident_id")
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		// Build WHERE clause dynamically
		whereClause := ""
		args := []interface{}{}

		if residentId != "" {
			whereClause += "p.resident_id = ?"
			args = append(args, residentId)
		}

		if startDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "p.payment_date >= ?"
			args = append(args, startDate)
		}

		if endDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "p.payment_date <= ?"
			args = append(args, endDate)
		}

		// Build full SQL query
		sqlQuery := `
			SELECT p.id, r.name, r.unit, p.amount, p.description, p.payment_date
			FROM payments p
			JOIN residents r ON p.resident_id = r.id
		`

		if whereClause != "" {
			sqlQuery += " WHERE " + whereClause
		}

		sqlQuery += " ORDER BY p.payment_date DESC"

		rows, err := db.Query(sqlQuery, args...)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		// Set headers for CSV download
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=payments_report_%s.csv",
			time.Now().Format("2006-01-02")))

		// Write CSV header
		fmt.Fprintf(w, "ID,Resident,Unit,Amount,Description,Date\n")

		// Write data rows
		for rows.Next() {
			var id int
			var name, unit, description, date string
			var amount float64

			if err := rows.Scan(&id, &name, &unit, &amount, &description, &date); err != nil {
				log.Printf("Error scanning payment row: %v", err)
				continue
			}

			// Escape description field for CSV (handle commas and quotes)
			if strings.Contains(description, ",") || strings.Contains(description, "\"") {
				description = "\"" + strings.ReplaceAll(description, "\"", "\"\"") + "\""
			}

			fmt.Fprintf(w, "%d,%s,%s,%.2f,%s,%s\n", id, name, unit, amount, description, date)
		}
	}
}

// Export expenses report as CSV
func exportExpensesReport(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters for filtering
		category := r.URL.Query().Get("category")
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		// Build WHERE clause dynamically
		whereClause := ""
		args := []interface{}{}

		if category != "" {
			whereClause += "category = ?"
			args = append(args, category)
		}

		if startDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "expense_date >= ?"
			args = append(args, startDate)
		}

		if endDate != "" {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "expense_date <= ?"
			args = append(args, endDate)
		}

		// Build full SQL query
		sqlQuery := "SELECT id, amount, description, expense_date, category FROM expenses"

		if whereClause != "" {
			sqlQuery += " WHERE " + whereClause
		}

		sqlQuery += " ORDER BY expense_date DESC"

		rows, err := db.Query(sqlQuery, args...)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		// Set headers for CSV download
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=expenses_report_%s.csv",
			time.Now().Format("2006-01-02")))

		// Write CSV header
		fmt.Fprintf(w, "ID,Amount,Description,Date,Category\n")

		// Write data rows
		for rows.Next() {
			var id int
			var description, date, category string
			var amount float64

			if err := rows.Scan(&id, &amount, &description, &date, &category); err != nil {
				log.Printf("Error scanning expense row: %v", err)
				continue
			}

			// Escape description field for CSV (handle commas and quotes)
			if strings.Contains(description, ",") || strings.Contains(description, "\"") {
				description = "\"" + strings.ReplaceAll(description, "\"", "\"\"") + "\""
			}

			fmt.Fprintf(w, "%d,%.2f,%s,%s,%s\n", id, amount, description, date, category)
		}
	}
}

// insertSampleData adds sample residents, payments, and expenses to the database
func insertSampleData(db *sql.DB) error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Clear existing data
	_, err = tx.Exec("DELETE FROM payments")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM expenses")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM residents")
	if err != nil {
		return err
	}

	// Insert sample residents
	residents := []struct {
		name    string
		unit    string
		contact string
		email   string
	}{
		{"John Smith", "101", "555-123-4567", "john.smith@example.com"},
		{"Jane Doe", "102", "555-234-5678", "jane.doe@example.com"},
		{"Robert Johnson", "201", "555-345-6789", "robert.j@example.com"},
		{"Maria Garcia", "202", "555-456-7890", "maria.g@example.com"},
		{"James Wilson", "301", "555-567-8901", "james.w@example.com"},
	}

	stmt, err := tx.Prepare("INSERT INTO residents(name, unit, contact, email) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	residentIDs := make([]int64, len(residents))
	for i, r := range residents {
		result, err := stmt.Exec(r.name, r.unit, r.contact, r.email)
		if err != nil {
			return err
		}
		residentIDs[i], err = result.LastInsertId()
		if err != nil {
			return err
		}
	}

	// Insert sample payments
	payments := []struct {
		residentIndex int
		amount        float64
		description   string
		date          string
	}{
		{0, 500.00, "Monthly maintenance fee", "2023-05-01"},
		{1, 500.00, "Monthly maintenance fee", "2023-05-02"},
		{2, 500.00, "Monthly maintenance fee", "2023-05-03"},
		{3, 500.00, "Monthly maintenance fee", "2023-05-05"},
		{4, 500.00, "Monthly maintenance fee", "2023-05-07"},
		{0, 500.00, "Monthly maintenance fee", "2023-06-01"},
		{1, 500.00, "Monthly maintenance fee", "2023-06-02"},
		{2, 500.00, "Monthly maintenance fee", "2023-06-04"},
	}

	stmt, err = tx.Prepare("INSERT INTO payments(resident_id, amount, description, payment_date) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range payments {
		_, err := stmt.Exec(residentIDs[p.residentIndex], p.amount, p.description, p.date)
		if err != nil {
			return err
		}
	}

	// Insert sample expenses
	expenses := []struct {
		amount      float64
		description string
		category    string
		date        string
	}{
		{1200.00, "Building cleaning", "Cleaning", "2023-05-15"},
		{350.50, "Elevator maintenance", "Maintenance", "2023-05-20"},
		{750.75, "Water bill", "Utilities", "2023-05-25"},
		{825.25, "Electricity bill", "Utilities", "2023-05-25"},
		{125.00, "Garden maintenance", "Maintenance", "2023-06-05"},
		{950.00, "Insurance premium", "Insurance", "2023-06-10"},
		{500.00, "Parking lot repair", "Maintenance", "2023-06-15"},
	}

	stmt, err = tx.Prepare("INSERT INTO expenses(amount, description, category, expense_date) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range expenses {
		_, err := stmt.Exec(e.amount, e.description, e.category, e.date)
		if err != nil {
			return err
		}
	}

	return nil
}
