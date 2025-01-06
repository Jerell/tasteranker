package types

import (
	"time"
)
type Item struct {
    ID            int       `json:"id"`
    TypeID        int       `json:"type_id"`
    Name          string    `json:"name"`
    CreatedAt     time.Time `json:"created_at"`
    LastUpdatedAt time.Time `json:"last_updated_at"`
}

type ItemType struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type ItemStore interface {
    CreateItem(item *Item) error
    GetItem(id int) (*Item, error)
    GetItemsByType(typeID int) ([]Item, error)
    UpdateItem(item *Item) error
}
