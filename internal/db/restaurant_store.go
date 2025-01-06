package db

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "github.com/Jerell/tasteranker/internal/types"
)

type PostgresRestaurantStore struct {
    db *sql.DB
}

func NewRestaurantStore(db *sql.DB) *PostgresRestaurantStore {
    return &PostgresRestaurantStore{db: db}
}

func (s *PostgresRestaurantStore) CreateChain(chain *types.RestaurantChain) error {
    query := `
        INSERT INTO restaurant_chains (
            name, website, description, founded_year
        ) VALUES ($1, $2, $3, $4)
        RETURNING id, created_at`

    err := s.db.QueryRow(
        query,
        chain.Name,
        chain.Website,
        chain.Description,
    ).Scan(&chain.ID, &chain.CreatedAt)

    if err != nil {
        return fmt.Errorf("failed to create restaurant chain: %v", err)
    }
    return nil
}

func (s *PostgresRestaurantStore) GetChainByID(id int) (*types.RestaurantChain, error) {
    chain := &types.RestaurantChain{}
    query := `
        SELECT id, name, website, description, founded_year, created_at
        FROM restaurant_chains
        WHERE id = $1`

    err := s.db.QueryRow(query, id).Scan(
        &chain.ID,
        &chain.Name,
        &chain.Website,
        &chain.Description,
        &chain.CreatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get restaurant chain: %v", err)
    }
    return chain, nil
}

func (s *PostgresRestaurantStore) CreateRestaurant(item *types.Item, metadata *types.RestaurantMetadata) error {
    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer tx.Rollback()

    query := `
        INSERT INTO items (type_id, name)
        VALUES ($1, $2)
        RETURNING id, created_at, last_updated_at`

    err = tx.QueryRow(query, item.TypeID, item.Name).Scan(
        &item.ID,
        &item.CreatedAt,
        &item.LastUpdatedAt,
    )
    if err != nil {
        return fmt.Errorf("failed to create item: %v", err)
    }

    metadata.ItemID = item.ID

    query = `
        INSERT INTO restaurant_metadata (
            item_id, chain_id, price_range,
            latitude, longitude, address, operating_hours,
            website, phone, google_place_id, google_maps_uri
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

    _, err = tx.Exec(
        query,
        metadata.ItemID,
        metadata.ChainID,
        metadata.PriceRange,
        metadata.Latitude,
        metadata.Longitude,
        metadata.Address,
        metadata.OperatingHours,
        metadata.Website,
        metadata.Phone,
        metadata.GooglePlaceID,
        metadata.GoogleMapsUri,
    )
    if err != nil {
        return fmt.Errorf("failed to create restaurant metadata: %v", err)
    }

    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }

    return nil
}

func (s *PostgresRestaurantStore) GetLocationsByChainID(chainID int) ([]types.RestaurantMetadata, error) {
    query := `
        SELECT 
            item_id, chain_id, price_range,
            latitude, longitude, address, operating_hours,
            website, phone, google_place_id, google_maps_uri
        FROM restaurant_metadata
        WHERE chain_id = $1
        ORDER BY item_id`

    rows, err := s.db.Query(query, chainID)
    if err != nil {
        return nil, fmt.Errorf("failed to query locations: %v", err)
    }
    defer rows.Close()

    var locations []types.RestaurantMetadata
    for rows.Next() {
        var loc types.RestaurantMetadata
        err := rows.Scan(
            &loc.ItemID,
            &loc.ChainID,
            &loc.PriceRange,
            &loc.Latitude,
            &loc.Longitude,
            &loc.Address,
            &loc.OperatingHours,
            &loc.Website,
            &loc.Phone,
            &loc.GooglePlaceID,
            &loc.GoogleMapsUri,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan location row: %v", err)
        }
        locations = append(locations, loc)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating location rows: %v", err)
    }

    return locations, nil
}

func (s *PostgresRestaurantStore) GetNearbyLocations(lat, lon float64, radiusKm float64) ([]types.RestaurantMetadata, error) {
    query := `
        SELECT 
            item_id, chain_id, price_range,
            latitude, longitude, address, operating_hours,
            website, phone, google_place_id, google_maps_uri
        FROM restaurant_metadata
        WHERE (
            6371 * acos(
                cos(radians($1)) *
                cos(radians(latitude)) *
                cos(radians(longitude) - radians($2)) +
                sin(radians($1)) *
                sin(radians(latitude))
            )
        ) <= $3
        ORDER BY item_id`

    rows, err := s.db.Query(query, lat, lon, radiusKm)
    if err != nil {
        return nil, fmt.Errorf("failed to query nearby locations: %v", err)
    }
    defer rows.Close()

    var locations []types.RestaurantMetadata
    for rows.Next() {
        var loc types.RestaurantMetadata
        err := rows.Scan(
            &loc.ItemID,
            &loc.ChainID,
            &loc.PriceRange,
            &loc.Latitude,
            &loc.Longitude,
            &loc.Address,
            &loc.OperatingHours,
            &loc.Website,
            &loc.Phone,
            &loc.GooglePlaceID,
            &loc.GoogleMapsUri,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan nearby location row: %v", err)
        }
        locations = append(locations, loc)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating nearby location rows: %v", err)
    }

    return locations, nil
}

func getPriceLevel(startPrice, endPrice int64) int {
    avgPrice := (startPrice + endPrice) / 2
    switch {
    case avgPrice <= 10:
        return 1  // Inexpensive
    case avgPrice <= 30:
        return 2  // Moderate
    case avgPrice <= 60:
        return 3  // Expensive
    default:
        return 4  // Very Expensive
    }
}

func (s *PostgresRestaurantStore) UpdateLocationFromGooglePlaces(placeID string, data *types.GooglePlacesResponse) error {
    query := `
        UPDATE restaurant_metadata
        SET 
            latitude = $1,
            longitude = $2,
            address = $3,
            operating_hours = $4,
            website = $5,
            phone = $6,
            price_range = $7,
            google_maps_uri = $8
        WHERE google_place_id = $9`

    priceLevel := getPriceLevel(
        data.PriceRange.StartPrice.Units,
        data.PriceRange.EndPrice.Units,
    )

    result, err := s.db.Exec(
        query,
        data.Location.Latitude,
        data.Location.Longitude,
        data.FormattedAddress,
        data.RegularOpeningHours,
        data.Website,
        data.InternationalPhoneNumber,
        priceLevel,
        data.GoogleMapsUri,
        placeID,
    )

    if err != nil {
        return fmt.Errorf("failed to update location from Google Places: %v", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking rows affected: %v", err)
    }

    if rows == 0 {
        return fmt.Errorf("no location found with Google Place ID: %s", placeID)
    }

    return nil
}

func (s *PostgresRestaurantStore) CreateMatchup(item1ID int, item2ID int, userID int, context json.RawMessage) (int, error) {
    query := `
        INSERT INTO matchups (item1_id, item2_id, user_id, context)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

    var matchupID int
    err := s.db.QueryRow(query, item1ID, item2ID, userID, context).Scan(&matchupID)
    if err != nil {
        return 0, fmt.Errorf("failed to create matchup: %v", err)
    }

    return matchupID, nil
}

func (s *PostgresRestaurantStore) RecordMatchupResult(matchupID int, winnerID int) error {
    query := `
        UPDATE matchups
        SET winner_id = $1
        WHERE id = $2`

    result, err := s.db.Exec(query, winnerID, matchupID)
    if err != nil {
        return fmt.Errorf("failed to record matchup result: %v", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking rows affected: %v", err)
    }

    if rows == 0 {
        return fmt.Errorf("no matchup found with ID: %d", matchupID)
    }

    return nil
}
