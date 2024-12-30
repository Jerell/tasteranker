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
        chain.FoundedYear,
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
        &chain.FoundedYear,
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

func (s *PostgresRestaurantStore) CreateLocation(metadata *types.RestaurantMetadata) error {
    query := `
        INSERT INTO restaurant_metadata (
            chain_id, cuisine_type, price_range,
            latitude, longitude, address, operating_hours,
            website, phone, google_place_id, rating,
            user_ratings_count
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING item_id`

    err := s.db.QueryRow(
        query,
        metadata.ChainID,
        metadata.CuisineType,
        metadata.PriceRange,
        metadata.Latitude,
        metadata.Longitude,
        metadata.Address,
        metadata.OperatingHours,
        metadata.Website,
        metadata.Phone,
        metadata.GooglePlaceID,
        metadata.Rating,
        metadata.UserRatingsCount,
    ).Scan(&metadata.ItemID)

    if err != nil {
        return fmt.Errorf("failed to create restaurant location: %v", err)
    }
    return nil
}

func (s *PostgresRestaurantStore) GetLocationsByChainID(chainID int) ([]types.RestaurantMetadata, error) {
    query := `
        SELECT 
            item_id, chain_id, cuisine_type, price_range,
            latitude, longitude, address, operating_hours,
            website, phone, google_place_id, rating,
            user_ratings_count
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
            &loc.CuisineType,
            &loc.PriceRange,
            &loc.Latitude,
            &loc.Longitude,
            &loc.Address,
            &loc.OperatingHours,
            &loc.Website,
            &loc.Phone,
            &loc.GooglePlaceID,
            &loc.Rating,
            &loc.UserRatingsCount,
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
            item_id, chain_id, cuisine_type, price_range,
            latitude, longitude, address, operating_hours,
            website, phone, google_place_id, rating,
            user_ratings_count
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
        ORDER BY rating DESC, user_ratings_count DESC`

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
            &loc.CuisineType,
            &loc.PriceRange,
            &loc.Latitude,
            &loc.Longitude,
            &loc.Address,
            &loc.OperatingHours,
            &loc.Website,
            &loc.Phone,
            &loc.GooglePlaceID,
            &loc.Rating,
            &loc.UserRatingsCount,
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
            rating = $7,
            user_ratings_count = $8,
            price_range = $9
        WHERE google_place_id = $10`

    result, err := s.db.Exec(
        query,
        data.Geometry.Location.Lat,
        data.Geometry.Location.Lng,
        data.FormattedAddress,
        data.OpeningHours,
        data.Website,
        data.InternationalPhoneNumber,
        data.Rating,
        data.UserRatings,
        data.PriceLevel,
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
