package eventstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// SQLiteEventStore is a SQLite-based event store implementation
type SQLiteEventStore struct {
	db     *sql.DB
	mu     sync.RWMutex
	logger logging.Logger
}

// NewSQLiteEventStore creates a new SQLite event store
func NewSQLiteEventStore(dbPath string) (*SQLiteEventStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteEventStore{
		db:     db,
		logger: logging.WithComponent("eventstore.sqlite"),
	}

	// Create tables if they don't exist
	if err := store.createTables(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to create tables: %w, also failed to close db: %w", err, closeErr)
		}
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// createTables creates the necessary database tables
func (s *SQLiteEventStore) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		aggregate_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		event_data TEXT NOT NULL,
		event_version INTEGER NOT NULL,
		occurred_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_aggregate_id ON events(aggregate_id);
	CREATE INDEX IF NOT EXISTS idx_occurred_at ON events(occurred_at);

	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		aggregate_id TEXT NOT NULL UNIQUE,
		snapshot_data TEXT NOT NULL,
		version INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := s.db.Exec(query)
	return err
}

// Save saves events for an aggregate with optimistic concurrency control
func (s *SQLiteEventStore) Save(aggregateID string, events []events.DomainEvent, expectedVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// Log error but don't return it since we're already handling the main error
			_ = err
		}
	}()

	// Check current version
	var currentVersion int
	err = tx.QueryRow(
		"SELECT COALESCE(MAX(event_version), 0) FROM events WHERE aggregate_id = ?",
		aggregateID,
	).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion != expectedVersion {
		return fmt.Errorf("concurrency conflict: expected version %d, but current version is %d", expectedVersion, currentVersion)
	}

	// Insert events
	stmt, err := tx.Prepare(`
		INSERT INTO events (aggregate_id, event_type, event_data, event_version, occurred_at)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			// Log error but don't return it
			_ = err
		}
	}()

	for _, event := range events {
		eventData, err := s.serializeEvent(event)
		if err != nil {
			return fmt.Errorf("failed to serialize event: %w", err)
		}

		_, err = stmt.Exec(
			event.AggregateID(),
			event.EventType(),
			eventData,
			event.EventVersion(),
			event.Timestamp().UTC(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Saved events",
		logging.String("aggregate_id", aggregateID),
		logging.Int("event_count", len(events)))

	return nil
}

// GetEvents retrieves all events for an aggregate
func (s *SQLiteEventStore) GetEvents(aggregateID string) ([]events.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT event_type, event_data, event_version, occurred_at
		FROM events
		WHERE aggregate_id = ?
		ORDER BY event_version ASC
	`, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't return it
			_ = err
		}
	}()

	var result []events.DomainEvent
	for rows.Next() {
		var eventType, eventData string
		var version int
		var occurredAt time.Time

		if err := rows.Scan(&eventType, &eventData, &version, &occurredAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		event, err := s.deserializeEvent(aggregateID, eventType, eventData, version, occurredAt)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event: %w", err)
		}

		result = append(result, event)
	}

	return result, nil
}

// GetEventsFromVersion retrieves events starting from a specific version
func (s *SQLiteEventStore) GetEventsFromVersion(aggregateID string, fromVersion int) ([]events.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT event_type, event_data, event_version, occurred_at
		FROM events
		WHERE aggregate_id = ? AND event_version > ?
		ORDER BY event_version ASC
	`, aggregateID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't return it
			_ = err
		}
	}()

	var result []events.DomainEvent
	for rows.Next() {
		var eventType, eventData string
		var version int
		var occurredAt time.Time

		if err := rows.Scan(&eventType, &eventData, &version, &occurredAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		event, err := s.deserializeEvent(aggregateID, eventType, eventData, version, occurredAt)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event: %w", err)
		}

		result = append(result, event)
	}

	return result, nil
}

// GetAllEvents returns all events in the store (for projections)
func (s *SQLiteEventStore) GetAllEvents() ([]events.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT aggregate_id, event_type, event_data, event_version, occurred_at
		FROM events
		ORDER BY occurred_at ASC, event_version ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all events: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't return it
			_ = err
		}
	}()

	var result []events.DomainEvent
	for rows.Next() {
		var aggregateID, eventType, eventData string
		var version int
		var occurredAt time.Time

		if err := rows.Scan(&aggregateID, &eventType, &eventData, &version, &occurredAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		event, err := s.deserializeEvent(aggregateID, eventType, eventData, version, occurredAt)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event: %w", err)
		}

		result = append(result, event)
	}

	return result, nil
}

// Close closes the database connection
func (s *SQLiteEventStore) Close() error {
	return s.db.Close()
}

// serializeEvent serializes an event to JSON
func (s *SQLiteEventStore) serializeEvent(event events.DomainEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// deserializeEvent deserializes an event from JSON
func (s *SQLiteEventStore) deserializeEvent(aggregateID, eventType, eventData string, version int, occurredAt time.Time) (events.DomainEvent, error) {
	// This is a simplified version - in production, you'd have a registry of event types
	// and proper deserialization logic for each event type
	
	// For now, we'll create a generic event wrapper
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(eventData), &data); err != nil {
		return nil, err
	}

	// Return a generic event - in production, you'd reconstruct the specific event type
	return &GenericEvent{
		aggregateID: aggregateID,
		eventType:   eventType,
		version:     version,
		timestamp:   occurredAt,
		data:        data,
	}, nil
}

// GenericEvent is a generic event implementation for deserialization
type GenericEvent struct {
	aggregateID string
	eventType   string
	version     int
	timestamp   time.Time
	data        map[string]interface{}
}

func (e *GenericEvent) AggregateID() string     { return e.aggregateID }
func (e *GenericEvent) EventType() string       { return e.eventType }
func (e *GenericEvent) EventVersion() int       { return e.version }
func (e *GenericEvent) Timestamp() time.Time    { return e.timestamp }
func (e *GenericEvent) Data() map[string]interface{} { return e.data }

// Ensure SQLiteEventStore implements EventStore
var _ EventStore = (*SQLiteEventStore)(nil)