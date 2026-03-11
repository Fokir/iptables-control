package traffic

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// InsertRaw inserts a raw traffic data point.
func (r *Repository) InsertRaw(nodeID *int64, interfaceName string, bytesIn, bytesOut int64, collectedAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO traffic_raw (node_id, interface_name, bytes_in, bytes_out, collected_at) VALUES (?, ?, ?, ?, ?)",
		nodeID, nilIfEmpty(interfaceName), bytesIn, bytesOut, collectedAt,
	)
	return err
}

// QueryRaw returns raw data points within a time range.
func (r *Repository) QueryRaw(nodeID *int64, interfaceName string, from, to time.Time) ([]TrafficPoint, error) {
	var rows *sql.Rows
	var err error

	if nodeID != nil {
		rows, err = r.db.Query(
			"SELECT collected_at, bytes_in, bytes_out FROM traffic_raw WHERE node_id = ? AND collected_at BETWEEN ? AND ? ORDER BY collected_at",
			*nodeID, from, to,
		)
	} else {
		rows, err = r.db.Query(
			"SELECT collected_at, bytes_in, bytes_out FROM traffic_raw WHERE interface_name = ? AND collected_at BETWEEN ? AND ? ORDER BY collected_at",
			interfaceName, from, to,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPoints(rows)
}

// QueryAggregated returns aggregated data points for a given bucket size.
func (r *Repository) QueryAggregated(nodeID *int64, interfaceName string, from, to time.Time, bucketSeconds int) ([]TrafficPoint, error) {
	var rows *sql.Rows
	var err error

	if nodeID != nil {
		rows, err = r.db.Query(
			"SELECT bucket_start, bytes_in, bytes_out FROM traffic_agg WHERE node_id = ? AND bucket_seconds = ? AND bucket_start BETWEEN ? AND ? ORDER BY bucket_start",
			*nodeID, bucketSeconds, from, to,
		)
	} else {
		rows, err = r.db.Query(
			"SELECT bucket_start, bytes_in, bytes_out FROM traffic_agg WHERE interface_name = ? AND bucket_seconds = ? AND bucket_start BETWEEN ? AND ? ORDER BY bucket_start",
			interfaceName, bucketSeconds, from, to,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPoints(rows)
}

// Aggregate rolls up raw data older than cutoff into buckets of targetSeconds.
func (r *Repository) AggregateRaw(cutoff time.Time, targetSeconds int) error {
	_, err := r.db.Exec(`
		INSERT INTO traffic_agg (node_id, interface_name, bytes_in, bytes_out, bucket_start, bucket_seconds, sample_count)
		SELECT node_id, interface_name,
			SUM(bytes_in), SUM(bytes_out),
			datetime((strftime('%s', collected_at) / ?) * ?, 'unixepoch'),
			?,
			COUNT(*)
		FROM traffic_raw
		WHERE collected_at < ?
		GROUP BY node_id, interface_name, (strftime('%s', collected_at) / ?)
	`, targetSeconds, targetSeconds, targetSeconds, cutoff, targetSeconds)
	return err
}

// PurgeRaw deletes raw data older than cutoff.
func (r *Repository) PurgeRaw(cutoff time.Time) error {
	_, err := r.db.Exec("DELETE FROM traffic_raw WHERE collected_at < ?", cutoff)
	return err
}

// AggregateAgg rolls up aggregated data from one bucket size to a larger one.
func (r *Repository) AggregateAgg(sourceBucket, targetBucket int, cutoff time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO traffic_agg (node_id, interface_name, bytes_in, bytes_out, bucket_start, bucket_seconds, sample_count)
		SELECT node_id, interface_name,
			SUM(bytes_in), SUM(bytes_out),
			datetime((strftime('%s', bucket_start) / ?) * ?, 'unixepoch'),
			?,
			SUM(sample_count)
		FROM traffic_agg
		WHERE bucket_seconds = ? AND bucket_start < ?
		GROUP BY node_id, interface_name, (strftime('%s', bucket_start) / ?)
	`, targetBucket, targetBucket, targetBucket, sourceBucket, cutoff, targetBucket)
	return err
}

// PurgeAgg deletes aggregated data older than cutoff for a specific bucket size.
func (r *Repository) PurgeAgg(bucketSeconds int, cutoff time.Time) error {
	_, err := r.db.Exec("DELETE FROM traffic_agg WHERE bucket_seconds = ? AND bucket_start < ?", bucketSeconds, cutoff)
	return err
}

// GetInterfaces returns distinct interface names from raw data.
func (r *Repository) GetInterfaces() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT interface_name FROM traffic_raw WHERE interface_name IS NOT NULL ORDER BY interface_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func scanPoints(rows *sql.Rows) ([]TrafficPoint, error) {
	var points []TrafficPoint
	for rows.Next() {
		var p TrafficPoint
		if err := rows.Scan(&p.Time, &p.BytesIn, &p.BytesOut); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if points == nil {
		points = []TrafficPoint{}
	}
	return points, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
