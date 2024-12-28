CREATE EXTENSION IF NOT EXISTS cube;
CREATE EXTENSION IF NOT EXISTS earthdistance;

CREATE TABLE restaurant_chains (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    website VARCHAR(255),
    description TEXT,
    founded_year INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE restaurant_metadata (
    item_id INTEGER PRIMARY KEY REFERENCES items(id),
    chain_id INTEGER REFERENCES restaurant_chains(id),
    cuisine_type VARCHAR(100)[],
    price_range INTEGER CHECK (price_range BETWEEN 1 AND 4),
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6),
    address TEXT,
    operating_hours JSONB,
    website VARCHAR(255),
    phone VARCHAR(50)
);

CREATE INDEX idx_restaurant_metadata_chain ON restaurant_metadata(chain_id);
CREATE INDEX idx_restaurant_metadata_cuisine ON restaurant_metadata USING gin(cuisine_type);
CREATE INDEX idx_restaurant_metadata_location ON restaurant_metadata USING gist (point(longitude, latitude));
