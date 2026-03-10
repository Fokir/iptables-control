package audit

import "time"

type LogEntry struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"userId"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	EntityType string    `json:"entityType"`
	EntityID   *int64    `json:"entityId"`
	Details    string    `json:"details,omitempty"`
	IPAddress  string    `json:"ipAddress"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ListParams struct {
	Page       int
	Limit      int
	Action     string
	EntityType string
}

type ListResult struct {
	Entries []LogEntry `json:"entries"`
	Total   int        `json:"total"`
	Page    int        `json:"page"`
	Limit   int        `json:"limit"`
}
