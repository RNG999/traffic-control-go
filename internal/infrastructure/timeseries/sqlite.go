package timeseries

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteTimeSeriesStore implements TimeSeriesStore using SQLite
type SQLiteTimeSeriesStore struct {
	db     *sql.DB
	logger logging.Logger
}

// NewSQLiteTimeSeriesStore creates a new SQLite-based time-series store
func NewSQLiteTimeSeriesStore(dbPath string) (*SQLiteTimeSeriesStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=memory")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	store := &SQLiteTimeSeriesStore{
		db:     db,
		logger: logging.WithComponent("timeseries.sqlite"),
	}

	if err := store.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	store.logger.Info("SQLite time-series store initialized", logging.String("db_path", dbPath))
	return store, nil
}

// createTables creates the necessary database tables
func (s *SQLiteTimeSeriesStore) createTables() error {
	schemas := []string{
		// Raw time-series data table
		`CREATE TABLE IF NOT EXISTS timeseries_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			qdisc_stats TEXT,
			class_stats TEXT,
			filter_stats TEXT,
			link_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Aggregated data tables for different intervals
		`CREATE TABLE IF NOT EXISTS aggregated_minute (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			qdisc_stats TEXT,
			class_stats TEXT,
			filter_stats TEXT,
			link_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS aggregated_hour (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			qdisc_stats TEXT,
			class_stats TEXT,
			filter_stats TEXT,
			link_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS aggregated_day (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			qdisc_stats TEXT,
			class_stats TEXT,
			filter_stats TEXT,
			link_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS aggregated_week (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			qdisc_stats TEXT,
			class_stats TEXT,
			filter_stats TEXT,
			link_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS aggregated_month (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			qdisc_stats TEXT,
			class_stats TEXT,
			filter_stats TEXT,
			link_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	// Create indexes for performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_timeseries_device_timestamp ON timeseries_data(device_name, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_timeseries_timestamp ON timeseries_data(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_agg_minute_device_timestamp ON aggregated_minute(device_name, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_agg_hour_device_timestamp ON aggregated_hour(device_name, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_agg_day_device_timestamp ON aggregated_day(device_name, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_agg_week_device_timestamp ON aggregated_week(device_name, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_agg_month_device_timestamp ON aggregated_month(device_name, timestamp)`,
	}

	// Execute schema creation
	for _, schema := range schemas {
		if _, err := s.db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create indexes
	for _, index := range indexes {
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// Store saves statistics data with timestamp
func (s *SQLiteTimeSeriesStore) Store(ctx context.Context, deviceName tc.DeviceName, stats *TimeSeriesData) error {
	// Serialize stats to JSON
	qdiscJSON, err := json.Marshal(stats.QdiscStats)
	if err != nil {
		return fmt.Errorf("failed to marshal qdisc stats: %w", err)
	}

	classJSON, err := json.Marshal(stats.ClassStats)
	if err != nil {
		return fmt.Errorf("failed to marshal class stats: %w", err)
	}

	filterJSON, err := json.Marshal(stats.FilterStats)
	if err != nil {
		return fmt.Errorf("failed to marshal filter stats: %w", err)
	}

	linkJSON, err := json.Marshal(stats.LinkStats)
	if err != nil {
		return fmt.Errorf("failed to marshal link stats: %w", err)
	}

	// Insert into database
	query := `INSERT INTO timeseries_data (device_name, timestamp, qdisc_stats, class_stats, filter_stats, link_stats)
	          VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err = s.db.ExecContext(ctx, query, deviceName.String(), stats.Timestamp, string(qdiscJSON), string(classJSON), string(filterJSON), string(linkJSON))
	if err != nil {
		return fmt.Errorf("failed to insert time-series data: %w", err)
	}

	s.logger.Debug("Stored time-series data", logging.String("device", deviceName.String()), logging.String("timestamp", stats.Timestamp.String()))
	return nil
}

// Query retrieves statistics data for a time range
func (s *SQLiteTimeSeriesStore) Query(ctx context.Context, deviceName tc.DeviceName, start, end time.Time) ([]*TimeSeriesData, error) {
	query := `SELECT device_name, timestamp, qdisc_stats, class_stats, filter_stats, link_stats
	          FROM timeseries_data 
	          WHERE device_name = ? AND timestamp >= ? AND timestamp <= ?
	          ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, query, deviceName.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query time-series data: %w", err)
	}
	defer rows.Close()

	var results []*TimeSeriesData
	for rows.Next() {
		var data TimeSeriesData
		var qdiscJSON, classJSON, filterJSON, linkJSON string

		err := rows.Scan(&data.DeviceName, &data.Timestamp, &qdiscJSON, &classJSON, &filterJSON, &linkJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Deserialize JSON data
		if err := json.Unmarshal([]byte(qdiscJSON), &data.QdiscStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal qdisc stats: %w", err)
		}
		if err := json.Unmarshal([]byte(classJSON), &data.ClassStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal class stats: %w", err)
		}
		if err := json.Unmarshal([]byte(filterJSON), &data.FilterStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filter stats: %w", err)
		}
		if err := json.Unmarshal([]byte(linkJSON), &data.LinkStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal link stats: %w", err)
		}

		results = append(results, &data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	s.logger.Debug("Queried time-series data", logging.String("device", deviceName.String()), logging.String("start", start.String()), logging.String("end", end.String()), logging.Int("count", len(results)))
	return results, nil
}

// QueryAggregated retrieves aggregated statistics for a time range with specified interval
func (s *SQLiteTimeSeriesStore) QueryAggregated(ctx context.Context, deviceName tc.DeviceName, start, end time.Time, interval AggregationInterval) ([]*AggregatedData, error) {
	tableName := s.getAggregationTable(interval)
	if tableName == "" {
		return nil, fmt.Errorf("unsupported aggregation interval: %s", interval)
	}

	query := fmt.Sprintf(`SELECT device_name, timestamp, qdisc_stats, class_stats, filter_stats, link_stats
	                      FROM %s 
	                      WHERE device_name = ? AND timestamp >= ? AND timestamp <= ?
	                      ORDER BY timestamp ASC`, tableName)

	rows, err := s.db.QueryContext(ctx, query, deviceName.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated data: %w", err)
	}
	defer rows.Close()

	var results []*AggregatedData
	for rows.Next() {
		var data AggregatedData
		var qdiscJSON, classJSON, filterJSON, linkJSON string

		err := rows.Scan(&data.DeviceName, &data.Timestamp, &qdiscJSON, &classJSON, &filterJSON, &linkJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan aggregated row: %w", err)
		}

		data.Interval = interval

		// Deserialize JSON data
		if err := json.Unmarshal([]byte(qdiscJSON), &data.QdiscStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal aggregated qdisc stats: %w", err)
		}
		if err := json.Unmarshal([]byte(classJSON), &data.ClassStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal aggregated class stats: %w", err)
		}
		if err := json.Unmarshal([]byte(filterJSON), &data.FilterStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal aggregated filter stats: %w", err)
		}
		if err := json.Unmarshal([]byte(linkJSON), &data.LinkStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal aggregated link stats: %w", err)
		}

		results = append(results, &data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during aggregated row iteration: %w", err)
	}

	s.logger.Debug("Queried aggregated data", logging.String("device", deviceName.String()), logging.String("interval", string(interval)), logging.String("start", start.String()), logging.String("end", end.String()), logging.Int("count", len(results)))
	return results, nil
}

// Delete removes old data based on retention policy
func (s *SQLiteTimeSeriesStore) Delete(ctx context.Context, deviceName tc.DeviceName, before time.Time) error {
	tables := []string{
		"timeseries_data",
		"aggregated_minute",
		"aggregated_hour", 
		"aggregated_day",
		"aggregated_week",
		"aggregated_month",
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var totalDeleted int64
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE device_name = ? AND timestamp < ?", table)
		result, err := tx.ExecContext(ctx, query, deviceName.String(), before)
		if err != nil {
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
		
		deleted, _ := result.RowsAffected()
		totalDeleted += deleted
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit deletion transaction: %w", err)
	}

	s.logger.Info("Deleted old time-series data", logging.String("device", deviceName.String()), logging.String("before", before.String()), logging.Int64("total_deleted", totalDeleted))
	return nil
}

// StoreAggregated stores aggregated data in the appropriate table
func (s *SQLiteTimeSeriesStore) StoreAggregated(ctx context.Context, deviceName tc.DeviceName, data *AggregatedData) error {
	tableName := s.getAggregationTable(data.Interval)
	if tableName == "" {
		return fmt.Errorf("unsupported aggregation interval: %s", data.Interval)
	}

	// Serialize stats to JSON
	qdiscJSON, err := json.Marshal(data.QdiscStats)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregated qdisc stats: %w", err)
	}

	classJSON, err := json.Marshal(data.ClassStats)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregated class stats: %w", err)
	}

	filterJSON, err := json.Marshal(data.FilterStats)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregated filter stats: %w", err)
	}

	linkJSON, err := json.Marshal(data.LinkStats)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregated link stats: %w", err)
	}

	// Insert into appropriate aggregation table
	query := fmt.Sprintf(`INSERT OR REPLACE INTO %s (device_name, timestamp, qdisc_stats, class_stats, filter_stats, link_stats)
	                      VALUES (?, ?, ?, ?, ?, ?)`, tableName)
	
	_, err = s.db.ExecContext(ctx, query, deviceName.String(), data.Timestamp, string(qdiscJSON), string(classJSON), string(filterJSON), string(linkJSON))
	if err != nil {
		return fmt.Errorf("failed to insert aggregated data into %s: %w", tableName, err)
	}

	s.logger.Debug("Stored aggregated data", logging.String("device", deviceName.String()), logging.String("interval", string(data.Interval)), logging.String("timestamp", data.Timestamp.String()), logging.String("table", tableName))
	return nil
}

// Close closes the store and cleans up resources
func (s *SQLiteTimeSeriesStore) Close() error {
	if s.db != nil {
		s.logger.Info("Closing SQLite time-series store")
		return s.db.Close()
	}
	return nil
}

// getAggregationTable returns the table name for the given aggregation interval
func (s *SQLiteTimeSeriesStore) getAggregationTable(interval AggregationInterval) string {
	switch interval {
	case IntervalMinute:
		return "aggregated_minute"
	case IntervalHour:
		return "aggregated_hour"
	case IntervalDay:
		return "aggregated_day"
	case IntervalWeek:
		return "aggregated_week"
	case IntervalMonth:
		return "aggregated_month"
	default:
		return ""
	}
}

// GetDatabaseStats returns database statistics for monitoring
func (s *SQLiteTimeSeriesStore) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	tables := []string{
		"timeseries_data",
		"aggregated_minute",
		"aggregated_hour",
		"aggregated_day", 
		"aggregated_week",
		"aggregated_month",
	}

	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := s.db.QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to get count for %s: %w", table, err)
		}
		stats[table+"_count"] = count
	}

	// Get database file size
	var pageCount, pageSize int64
	err := s.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}
	
	err = s.db.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get page size: %w", err)
	}

	stats["database_size_bytes"] = pageCount * pageSize

	return stats, nil
}

// Vacuum performs database maintenance operations
func (s *SQLiteTimeSeriesStore) Vacuum(ctx context.Context) error {
	s.logger.Info("Starting database vacuum operation")
	
	_, err := s.db.ExecContext(ctx, "VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}
	
	s.logger.Info("Database vacuum completed")
	return nil
}