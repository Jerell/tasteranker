package db

import (
    "database/sql"
    "fmt"
    "github.com/Jerell/tasteranker/internal/types"
)

type itemStore struct {
    db *sql.DB
}

func NewItemStore(db *sql.DB) *itemStore {
    return &itemStore{db: db}
}

func (s *itemStore) CreateItem(item *types.Item) error {
    query := `
        INSERT INTO items (type_id, name)
        VALUES ($1, $2)
        RETURNING id, created_at, last_updated_at`

    return s.db.QueryRow(
        query,
        item.TypeID,
        item.Name,
    ).Scan(
        &item.ID,
        &item.CreatedAt,
        &item.LastUpdatedAt,
    )
}

func (s *itemStore) GetItem(id int) (*types.Item, error) {
    item := &types.Item{}
    query := `
        SELECT id, type_id, name, created_at, last_updated_at
        FROM items
        WHERE id = $1`

    err := s.db.QueryRow(query, id).Scan(
        &item.ID,
        &item.TypeID,
        &item.Name,
        &item.CreatedAt,
        &item.LastUpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("item not found: %v", err)
        }
        return nil, fmt.Errorf("error fetching item: %v", err)
    }
    return item, nil
}

func (s *itemStore) GetItemsByType(typeID int) ([]types.Item, error) {
    query := `
        SELECT id, type_id, name, created_at, last_updated_at
        FROM items
        WHERE type_id = $1
        ORDER BY name`

    rows, err := s.db.Query(query, typeID)
    if err != nil {
        return nil, fmt.Errorf("error querying items by type: %v", err)
    }
    defer rows.Close()

    var items []types.Item
    for rows.Next() {
        var item types.Item
        err := rows.Scan(
            &item.ID,
            &item.TypeID,
            &item.Name,
            &item.CreatedAt,
            &item.LastUpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning item row: %v", err)
        }
        items = append(items, item)
    }
    return items, rows.Err()
}

func (s *itemStore) UpdateItem(item *types.Item) error {
    query := `
        UPDATE items
        SET name = $1,
            last_updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING last_updated_at`

    return s.db.QueryRow(
        query,
        item.Name,
        item.ID,
    ).Scan(&item.LastUpdatedAt)
}
