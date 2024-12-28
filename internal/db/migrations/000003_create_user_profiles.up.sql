CREATE TABLE user_profiles (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    preferences JSONB,
    home_location_lat DECIMAL(9,6),
    home_location_lon DECIMAL(9,6),
    dietary_restrictions VARCHAR(50)[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
