package models

import (
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Operation represents a queued operation
type Operation struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Operation string    `json:"operation"`
	Payload   string    `json:"payload"`
	CreatedAt int64     `json:"created_at"`
	Status    string    `json:"status"`
	Attempts  int       `json:"attempts"`
}

// Queue manages operation persistence
type Queue struct {
	db *sql.DB
	mu sync.Mutex
}

// NewQueue creates a new queue with SQLite backend
func NewQueue(dbPath string) (*Queue, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Enable connection pooling
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	q := &Queue{db: db}

	// Initialize schema
	if err := q.initSchema(); err != nil {
		return nil, err
	}

	return q, nil
}

// initSchema creates the required tables
func (q *Queue) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS pending_operations (
		id TEXT PRIMARY KEY,
		account_id TEXT NOT NULL,
		operation TEXT NOT NULL,
		payload TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		status TEXT DEFAULT 'PENDING',
		attempts INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_pending_operations_account_id
	ON pending_operations(account_id);
	CREATE INDEX IF NOT EXISTS idx_pending_operations_created_at
	ON pending_operations(created_at);
	CREATE INDEX IF NOT EXISTS idx_pending_operations_status
	ON pending_operations(status);
	CREATE INDEX IF NOT EXISTS idx_pending_operations_status_created_at
	ON pending_operations(status, created_at);
	`

	_, err := q.db.Exec(schema)
	return err
}

// Enqueue adds a new operation to the queue
func (q *Queue) Enqueue(op *Operation) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if op.CreatedAt == 0 {
		op.CreatedAt = time.Now().Unix()
	}
	if op.Status == "" {
		op.Status = "PENDING"
	}

	query := `
	INSERT INTO pending_operations (id, account_id, operation, payload, created_at, status, attempts)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := q.db.Exec(query, op.ID, op.AccountID, op.Operation, op.Payload, op.CreatedAt, op.Status, op.Attempts)
	return err
}

// Dequeue retrieves the oldest pending operation (FIFO)
func (q *Queue) Dequeue() (*Operation, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	query := `
	SELECT id, account_id, operation, payload, created_at, status, attempts
	FROM pending_operations
	WHERE status = 'PENDING'
	ORDER BY created_at ASC
	LIMIT 1
	`

	row := q.db.QueryRow(query)

	var op Operation
	err := row.Scan(&op.ID, &op.AccountID, &op.Operation, &op.Payload, &op.CreatedAt, &op.Status, &op.Attempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &op, nil
}

// UpdateStatus updates the status of an operation
func (q *Queue) UpdateStatus(id, status string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	query := `UPDATE pending_operations SET status = ? WHERE id = ?`
	_, err := q.db.Exec(query, status, id)
	return err
}

// IncrementAttempts increments the attempt count
func (q *Queue) IncrementAttempts(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	query := `UPDATE pending_operations SET attempts = attempts + 1 WHERE id = ?`
	_, err := q.db.Exec(query, id)
	return err
}

// GetByID retrieves an operation by ID
func (q *Queue) GetByID(id string) (*Operation, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	query := `
	SELECT id, account_id, operation, payload, created_at, status, attempts
	FROM pending_operations
	WHERE id = ?
	`

	row := q.db.QueryRow(query, id)

	var op Operation
	err := row.Scan(&op.ID, &op.AccountID, &op.Operation, &op.Payload, &op.CreatedAt, &op.Status, &op.Attempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &op, nil
}

// ListByStatus retrieves all operations with a specific status
func (q *Queue) ListByStatus(status string) ([]*Operation, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	query := `
	SELECT id, account_id, operation, payload, created_at, status, attempts
	FROM pending_operations
	WHERE status = ?
	ORDER BY created_at ASC
	`

	rows, err := q.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var operations []*Operation
	for rows.Next() {
		var op Operation
		if err := rows.Scan(&op.ID, &op.AccountID, &op.Operation, &op.Payload, &op.CreatedAt, &op.Status, &op.Attempts); err != nil {
			return nil, err
		}
		operations = append(operations, &op)
	}

	return operations, rows.Err()
}

// GetQueueLength returns the number of pending operations
func (q *Queue) GetQueueLength() (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var count int
	query := `SELECT COUNT(*) FROM pending_operations WHERE status = 'PENDING'`
	err := q.db.QueryRow(query).Scan(&count)
	return count, err
}

// Close closes the database connection
func (q *Queue) Close() error {
	return q.db.Close()
}

// PayloadToMap converts a JSON payload string to a map
func PayloadToMap(payload string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(payload), &result)
	return result, err
}
