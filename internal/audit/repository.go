package audit

import (
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(entry *LogEntry) error {
	_, err := r.db.Exec(
		"INSERT INTO audit_logs (user_id, username, action, entity_type, entity_id, details, ip_address) VALUES (?, ?, ?, ?, ?, ?, ?)",
		entry.UserID, entry.Username, entry.Action, entry.EntityType, entry.EntityID, entry.Details, entry.IPAddress,
	)
	return err
}

func (r *Repository) List(params ListParams) (*ListResult, error) {
	var conditions []string
	var args []any

	if params.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, params.Action)
	}
	if params.EntityType != "" {
		conditions = append(conditions, "entity_type = ?")
		args = append(args, params.EntityType)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", where), args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	query := fmt.Sprintf(
		"SELECT id, user_id, username, action, entity_type, entity_id, details, ip_address, created_at FROM audit_logs %s ORDER BY created_at DESC LIMIT ? OFFSET ?",
		where,
	)
	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		var details sql.NullString
		if err := rows.Scan(&e.ID, &e.UserID, &e.Username, &e.Action, &e.EntityType, &e.EntityID, &details, &e.IPAddress, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Details = details.String
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []LogEntry{}
	}

	return &ListResult{
		Entries: entries,
		Total:   total,
		Page:    params.Page,
		Limit:   params.Limit,
	}, nil
}

func (r *Repository) Cleanup(days int) error {
	_, err := r.db.Exec(
		"DELETE FROM audit_logs WHERE created_at < datetime('now', ? || ' days')",
		fmt.Sprintf("-%d", days),
	)
	return err
}
