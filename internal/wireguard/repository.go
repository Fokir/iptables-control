package wireguard

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

func (r *Repository) GetAll() ([]Peer, error) {
	rows, err := r.db.Query("SELECT id, name, public_key, preshared_key, private_key, allowed_ips, address, dns, created_at FROM wireguard_peers ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []Peer
	for rows.Next() {
		var p Peer
		if err := rows.Scan(&p.ID, &p.Name, &p.PublicKey, &p.PresharedKey, &p.PrivateKey, &p.AllowedIPs, &p.Address, &p.DNS, &p.CreatedAt); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	if peers == nil {
		peers = []Peer{}
	}
	return peers, nil
}

func (r *Repository) GetByID(id int64) (*Peer, error) {
	var p Peer
	err := r.db.QueryRow(
		"SELECT id, name, public_key, preshared_key, private_key, allowed_ips, address, dns, created_at FROM wireguard_peers WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.PublicKey, &p.PresharedKey, &p.PrivateKey, &p.AllowedIPs, &p.Address, &p.DNS, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetByName(name string) (*Peer, error) {
	var p Peer
	err := r.db.QueryRow(
		"SELECT id, name, public_key, preshared_key, private_key, allowed_ips, address, dns, created_at FROM wireguard_peers WHERE name = ?",
		name,
	).Scan(&p.ID, &p.Name, &p.PublicKey, &p.PresharedKey, &p.PrivateKey, &p.AllowedIPs, &p.Address, &p.DNS, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Create(p *Peer) error {
	res, err := r.db.Exec(
		"INSERT INTO wireguard_peers (name, public_key, preshared_key, private_key, allowed_ips, address, dns) VALUES (?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.PublicKey, p.PresharedKey, p.PrivateKey, p.AllowedIPs, p.Address, p.DNS,
	)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	p.CreatedAt = time.Now()
	return nil
}

func (r *Repository) Update(p *Peer) error {
	_, err := r.db.Exec(
		"UPDATE wireguard_peers SET name = ?, allowed_ips = ?, dns = ? WHERE id = ?",
		p.Name, p.AllowedIPs, p.DNS, p.ID,
	)
	return err
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM wireguard_peers WHERE id = ?", id)
	return err
}
