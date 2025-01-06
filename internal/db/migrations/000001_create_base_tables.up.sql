CREATE EXTENSION IF NOT EXISTS cube;
CREATE EXTENSION IF NOT EXISTS earthdistance;

CREATE TABLE item_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    type_id INTEGER REFERENCES item_types(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    preferences JSONB,
    home_location_lat DECIMAL(9,6),
    home_location_lon DECIMAL(9,6),
    dietary_restrictions VARCHAR(50)[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE auth_providers (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(20) NOT NULL,
    provider_user_id VARCHAR(50) NOT NULL,
    user_profile_id INTEGER REFERENCES user_profiles(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id)
);

CREATE TABLE restaurant_chains (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    website VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE restaurant_metadata (
    item_id INTEGER PRIMARY KEY REFERENCES items(id),
    chain_id INTEGER REFERENCES restaurant_chains(id),
    price_range INTEGER CHECK (price_range BETWEEN 1 AND 4),
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6),
    address TEXT,
    operating_hours JSONB,
    website VARCHAR(255),
    phone VARCHAR(50),
    google_place_id VARCHAR(255) UNIQUE,
    google_maps_uri VARCHAR(255)
);

CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_by INTEGER REFERENCES user_profiles(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    event_time TIMESTAMP,
    status VARCHAR(50) DEFAULT 'planning'
);

CREATE TABLE group_members (
    group_id INTEGER REFERENCES groups(id),
    user_profile_id INTEGER REFERENCES user_profiles(id),
    search_radius_meters INTEGER,
    availability JSONB,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_profile_id)
);

CREATE TABLE matchups (
    id SERIAL PRIMARY KEY,
    item1_id INTEGER REFERENCES items(id),
    item2_id INTEGER REFERENCES items(id),
    winner_id INTEGER REFERENCES items(id),
    user_profile_id INTEGER REFERENCES user_profiles(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    context JSONB,
    UNIQUE(item1_id, item2_id, user_profile_id)
);

-- Add initial item type
INSERT INTO item_types (name) VALUES ('restaurant');

-- Indexes
CREATE INDEX idx_restaurant_metadata_chain ON restaurant_metadata(chain_id);
CREATE INDEX idx_restaurant_metadata_location ON restaurant_metadata USING gist (point(longitude, latitude));
CREATE INDEX idx_matchups_temporal ON matchups(created_at);
