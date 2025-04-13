# Condo Manager

A simple application to manage condo expenses, residents, and payments using Go, SQLite, and JavaScript.

## Features

- **Residents Management**: Add, edit, and delete residents with unit information
- **Payment Tracking**: Record and track payments from residents
- **Expense Management**: Track condo expenses by category
- **Dashboard**: Overview of residents, payments, and expenses
- **Search Functionality**: Quickly find residents, payments, and expenses with real-time search
- **Data Validation**: Input validation for all forms to ensure data integrity
- **Data Import/Export**: Export database to JSON and import from JSON files for backup and migration
- **Report Generation**: Export payments and expenses reports to CSV
- **Charts & Visualizations**: View payment trends and expense breakdown with interactive charts
- **Single Binary**: All resources are embedded in a single Go binary

## Technologies Used

- **Backend**: Go with SQLite database
- **Frontend**: HTML, CSS, JavaScript, Bootstrap 5
- **Libraries**:
  - `github.com/mattn/go-sqlite3` - SQLite driver
  - `github.com/gorilla/mux` - HTTP router
  - Chart.js - Data visualization

## Getting Started

### Prerequisites

- Go 1.16 or later (for Go embed feature)

### Building

```bash
# Clone the repository
git clone https://github.com/yourusername/condomngr.git
cd condomngr

# Build the application
go build -o condomngr

# Run the application
./condomngr
```

The application will start a web server on port 8080. Open your browser and navigate to `http://localhost:8080` to access the application.

### Loading Sample Data

To start the application with sample data (for demonstration or testing purposes), use the `-sample` flag:

```bash
./condomngr -sample
```

This will:
- Load 5 sample residents
- Add 8 sample payments
- Create 7 sample expenses

Note: Loading sample data will clear any existing data in the database.

## Advanced Features

### Database Export and Import

The application allows exporting and importing the entire database:

1. **Exporting**: Click the "Export Database" button on the Dashboard to download a JSON file with all data
2. **Importing**: Click the "Import Database" button and select a previously exported JSON file to restore data

### Report Generation

Generate and download reports in CSV format:

1. **Payments Report**: Click the "Export CSV" button on the Payments page
2. **Expenses Report**: Click the "Export CSV" button on the Expenses page

### Search and Filtering

Use the search boxes at the top of each section to quickly find:

- Residents by name, unit, contact, or email
- Payments by description or resident
- Expenses by description or category

## API Endpoints

### Residents

- `GET /api/residents` - Get all residents
- `POST /api/residents` - Create a new resident
- `GET /api/residents/{id}` - Get a specific resident
- `PUT /api/residents/{id}` - Update a resident
- `DELETE /api/residents/{id}` - Delete a resident

### Payments

- `GET /api/payments` - Get all payments
- `POST /api/payments` - Create a new payment
- `GET /api/payments/{id}` - Get a specific payment
- `PUT /api/payments/{id}` - Update a payment
- `DELETE /api/payments/{id}` - Delete a payment

### Expenses

- `GET /api/expenses` - Get all expenses
- `POST /api/expenses` - Create a new expense
- `GET /api/expenses/{id}` - Get a specific expense
- `PUT /api/expenses/{id}` - Update an expense
- `DELETE /api/expenses/{id}` - Delete an expense

### Data Import/Export

- `GET /api/export` - Export database as JSON
- `POST /api/import` - Import database from JSON

### Search

- `GET /api/search/residents?q={query}` - Search residents
- `GET /api/search/payments?q={query}` - Search payments
- `GET /api/search/expenses?q={query}` - Search expenses

### Reports

- `GET /api/reports/payments/export` - Export payments report as CSV
- `GET /api/reports/expenses/export` - Export expenses report as CSV

## Data Structure

### Residents
```json
{
  "id": 1,
  "name": "John Doe",
  "unit": "101",
  "contact": "555-123-4567",
  "email": "john.doe@example.com",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Payments
```json
{
  "id": 1,
  "resident_id": 1,
  "amount": 500.00,
  "description": "Monthly maintenance fee",
  "payment_date": "2023-01-15",
  "created_at": "2023-01-15T00:00:00Z"
}
```

### Expenses
```json
{
  "id": 1,
  "amount": 350.00,
  "description": "Building maintenance",
  "expense_date": "2023-01-10",
  "category": "Maintenance",
  "created_at": "2023-01-10T00:00:00Z"
}
```

## License

MIT 