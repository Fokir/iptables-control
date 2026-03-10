package auth

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

func (r *Repository) GetUserByUsername(username string) (*User, error) {
	var u User
	err := r.db.QueryRow(
		"SELECT id, username, password, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(username, hashedPassword string) (*User, error) {
	res, err := r.db.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?)",
		username, hashedPassword,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &User{ID: id, Username: username, CreatedAt: time.Now()}, nil
}

func (r *Repository) UserCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *Repository) CreateSession(session *Session) error {
	_, err := r.db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		session.ID, session.UserID, session.ExpiresAt,
	)
	return err
}

func (r *Repository) GetSession(id string) (*Session, error) {
	var s Session
	err := r.db.QueryRow(
		"SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = ? AND expires_at > ?",
		id, time.Now(),
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ExtendSession(id string, expiresAt time.Time) error {
	_, err := r.db.Exec("UPDATE sessions SET expires_at = ? WHERE id = ?", expiresAt, id)
	return err
}

func (r *Repository) DeleteSession(id string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

func (r *Repository) DeleteExpiredSessions() error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

func (r *Repository) GetUserByID(id int64) (*User, error) {
	var u User
	err := r.db.QueryRow(
		"SELECT id, username, password, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
