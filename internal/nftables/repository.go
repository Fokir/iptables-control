package nftables

import (
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]NatGroup, error) {
	rows, err := r.db.Query("SELECT id, name, enabled, target_ip, target_reverse_ip, destination_ip, created_at, updated_at FROM nat_groups ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []NatGroup
	for rows.Next() {
		var g NatGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Enabled, &g.TargetIP, &g.TargetReverseIP, &g.DestinationIP, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	for i := range groups {
		ports, err := r.GetPortsByGroupID(groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Ports = ports
	}

	return groups, nil
}

func (r *Repository) GetByID(id int64) (*NatGroup, error) {
	var g NatGroup
	err := r.db.QueryRow(
		"SELECT id, name, enabled, target_ip, target_reverse_ip, destination_ip, created_at, updated_at FROM nat_groups WHERE id = ?",
		id,
	).Scan(&g.ID, &g.Name, &g.Enabled, &g.TargetIP, &g.TargetReverseIP, &g.DestinationIP, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}

	ports, err := r.GetPortsByGroupID(g.ID)
	if err != nil {
		return nil, err
	}
	g.Ports = ports

	return &g, nil
}

func (r *Repository) Create(g *NatGroup) error {
	res, err := r.db.Exec(
		"INSERT INTO nat_groups (name, enabled, target_ip, target_reverse_ip, destination_ip) VALUES (?, ?, ?, ?, ?)",
		g.Name, g.Enabled, g.TargetIP, g.TargetReverseIP, g.DestinationIP,
	)
	if err != nil {
		return err
	}
	g.ID, _ = res.LastInsertId()
	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()
	return nil
}

func (r *Repository) Update(g *NatGroup) error {
	_, err := r.db.Exec(
		"UPDATE nat_groups SET name = ?, enabled = ?, target_ip = ?, target_reverse_ip = ?, destination_ip = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		g.Name, g.Enabled, g.TargetIP, g.TargetReverseIP, g.DestinationIP, g.ID,
	)
	return err
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM nat_groups WHERE id = ?", id)
	return err
}

func (r *Repository) GetPortsByGroupID(groupID int64) ([]NatPort, error) {
	rows, err := r.db.Query(
		"SELECT id, group_id, external_port, internal_port, protocol_tcp, protocol_udp, created_at FROM nat_ports WHERE group_id = ? ORDER BY id",
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []NatPort
	for rows.Next() {
		var p NatPort
		if err := rows.Scan(&p.ID, &p.GroupID, &p.ExternalPort, &p.InternalPort, &p.ProtocolTCP, &p.ProtocolUDP, &p.CreatedAt); err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	if ports == nil {
		ports = []NatPort{}
	}
	return ports, nil
}

func (r *Repository) ReplaceGroupPorts(groupID int64, ports []CreatePortRequest) ([]NatPort, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM nat_ports WHERE group_id = ?", groupID); err != nil {
		return nil, fmt.Errorf("delete old ports: %w", err)
	}

	var result []NatPort
	for _, p := range ports {
		res, err := tx.Exec(
			"INSERT INTO nat_ports (group_id, external_port, internal_port, protocol_tcp, protocol_udp) VALUES (?, ?, ?, ?, ?)",
			groupID, p.ExternalPort, p.InternalPort, p.ProtocolTCP, p.ProtocolUDP,
		)
		if err != nil {
			return nil, fmt.Errorf("insert port: %w", err)
		}
		id, _ := res.LastInsertId()
		result = append(result, NatPort{
			ID:           id,
			GroupID:      groupID,
			ExternalPort: p.ExternalPort,
			InternalPort: p.InternalPort,
			ProtocolTCP:  p.ProtocolTCP,
			ProtocolUDP:  p.ProtocolUDP,
			CreatedAt:    time.Now(),
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if result == nil {
		result = []NatPort{}
	}
	return result, nil
}

func (r *Repository) GetAllEnabled() ([]NatGroup, error) {
	rows, err := r.db.Query("SELECT id, name, enabled, target_ip, target_reverse_ip, destination_ip, created_at, updated_at FROM nat_groups WHERE enabled = 1 ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []NatGroup
	for rows.Next() {
		var g NatGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Enabled, &g.TargetIP, &g.TargetReverseIP, &g.DestinationIP, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	for i := range groups {
		ports, err := r.GetPortsByGroupID(groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Ports = ports
	}

	return groups, nil
}
