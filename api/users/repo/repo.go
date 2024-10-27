package userrepo

import (
    "database/sql"
)

type Repo struct {
    db *sql.DB
}

func (r *Repo) CreateUser(name, email string) error {
    _, err := r.db.Exec(`
        INSERT INTO users (email, name) VALUES ($1, $2)
    `, email, name)
    return err
}
