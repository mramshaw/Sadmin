package servers

import "database/sql"

// The Server entity is used to marshall/unmarshall JSON.
type Server struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetServer returns a single specified server.
func (s *Server) GetServer(db *sql.DB) error {

	stmt, err := db.Prepare("SELECT name FROM servers WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(s.ID).Scan(&s.Name)
}

// UpdateServer is used to modify a specific server.
func (s *Server) UpdateServer(db *sql.DB) (res sql.Result, err error) {

	stmt, err := db.Prepare("UPDATE servers SET name = ? WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return stmt.Exec(s.Name, s.ID)
}

// DeleteServer is used to delete a specific server.
func (s *Server) DeleteServer(db *sql.DB) (res sql.Result, err error) {

	stmt, err := db.Prepare("DELETE FROM servers WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return stmt.Exec(s.ID)
}

// CreateServer is used to create a single server.
func (s *Server) CreateServer(db *sql.DB) error {

	stmt, err := db.Prepare("INSERT INTO servers (name) VALUES(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(s.Name)
	if err != nil {
		return err
	}

	s.ID, err = res.LastInsertId()

	return err
}

// GetServers returns a collection of known servers.
func GetServers(db *sql.DB, start int, count int) ([]Server, error) {

	stmt, err := db.Prepare("SELECT id, name FROM servers ORDER BY name LIMIT ? OFFSET ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(count, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []Server{}
	for rows.Next() {
		var s Server
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, nil
}

// SearchServers returns a collection of servers matching the search criteria.
func SearchServers(db *sql.DB, start int, count int, name string) ([]Server, error) {

	stmt, err := db.Prepare("SELECT id, name FROM servers WHERE name LIKE ? ORDER BY name LIMIT ? OFFSET ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(name, count, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []Server{}
	for rows.Next() {
		var s Server
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, nil
}
