package types

import (
    "encoding/json"
)

type MatchupPair struct {
    ID      int     `json:"id"`
    Item1   Item    `json:"item1"`
    Item2   Item    `json:"item2"`
    Context Context `json:"context"`
}

type MatchupResult struct {
    UserID   int     `json:"user_id"`
    PairID   int     `json:"pair_id"`
    WinnerID *int    `json:"winner_id"` // nil for draw
    Context  Context `json:"context"`
}

type Context struct {
    Context json.RawMessage `json:"context,omitempty"`
}
