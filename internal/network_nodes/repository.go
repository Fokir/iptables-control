package network_nodes

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]NetworkNode, error) {
	rows, err := r.db.Query("SELECT id, name, ip, created_at FROM network_nodes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []NetworkNode
	for rows.Next() {
		var n NetworkNode
		if err := rows.Scan(&n.ID, &n.Name, &n.IP, &n.CreatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	if nodes == nil {
		nodes = []NetworkNode{}
	}
	return nodes, nil
}

func (r *Repository) Create(n *NetworkNode) error {
	res, err := r.db.Exec("INSERT INTO network_nodes (name, ip) VALUES (?, ?)", n.Name, n.IP)
	if err != nil {
		return err
	}
	n.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) Update(n *NetworkNode) error {
	_, err := r.db.Exec("UPDATE network_nodes SET name = ?, ip = ? WHERE id = ?", n.Name, n.IP, n.ID)
	return err
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM network_nodes WHERE id = ?", id)
	return err
}
