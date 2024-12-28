CREATE TABLE matchups (
    id SERIAL PRIMARY KEY,
    item1_id INTEGER REFERENCES items(id),
    item2_id INTEGER REFERENCES items(id),
    winner_id INTEGER REFERENCES items(id),
    user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    context JSONB,
    UNIQUE(item1_id, item2_id, user_id)
);

CREATE INDEX idx_matchups_temporal ON matchups(created_at);
