package db

import (
    "database/sql"
    "fmt"
    "github.com/lib/pq"
    "github.com/Jerell/tasteranker/internal/types"
)

type userStore struct {
    db *sql.DB
}

func NewUserStore(db *sql.DB) *userStore {
    return &userStore{db: db}
}

func (s *userStore) CreateProfile(profile *types.UserProfile) error {
    query := `
        INSERT INTO user_profiles (
            preferences, home_location_lat, home_location_lon, 
            dietary_restrictions
        ) VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at`

    return s.db.QueryRow(
        query,
        profile.Preferences,
        profile.HomeLocationLat,
        profile.HomeLocationLon,
        pq.Array(profile.DietaryRestrictions),
    ).Scan(
        &profile.ID,
        &profile.CreatedAt,
        &profile.UpdatedAt,
    )
}

func (s *userStore) GetProfile(id int) (*types.UserProfile, error) {
    profile := &types.UserProfile{}
    query := `
        SELECT id, preferences, home_location_lat, home_location_lon,
               dietary_restrictions, created_at, updated_at
        FROM user_profiles
        WHERE id = $1`

    err := s.db.QueryRow(query, id).Scan(
        &profile.ID,
        &profile.Preferences,
        &profile.HomeLocationLat,
        &profile.HomeLocationLon,
        pq.Array(&profile.DietaryRestrictions),
        &profile.CreatedAt,
        &profile.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("profile not found: %v", err)
        }
        return nil, fmt.Errorf("error fetching profile: %v", err)
    }
    return profile, nil
}

func (s *userStore) GetProfileByProviderID(provider, providerUserID string) (*types.UserProfile, error) {
    profile := &types.UserProfile{}
    query := `
        SELECT p.id, p.preferences, p.home_location_lat, p.home_location_lon,
               p.dietary_restrictions, p.created_at, p.updated_at
        FROM user_profiles p
        JOIN auth_providers a ON a.user_profile_id = p.id
        WHERE a.provider = $1 AND a.provider_user_id = $2`

    err := s.db.QueryRow(query, provider, providerUserID).Scan(
        &profile.ID,
        &profile.Preferences,
        &profile.HomeLocationLat,
        &profile.HomeLocationLon,
        pq.Array(&profile.DietaryRestrictions),
        &profile.CreatedAt,
        &profile.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // No profile found is not an error in this case
        }
        return nil, fmt.Errorf("error fetching profile by provider: %v", err)
    }
    return profile, nil
}

func (s *userStore) CreateAuthProvider(auth *types.AuthProvider) error {
    query := `
        INSERT INTO auth_providers (
            provider, provider_user_id, user_profile_id
        ) VALUES ($1, $2, $3)
        RETURNING id, created_at`

    return s.db.QueryRow(
        query,
        auth.Provider,
        auth.ProviderUserID,
        auth.UserProfileID,
    ).Scan(&auth.ID, &auth.CreatedAt)
}

func (s *userStore) UpdateProfile(profile *types.UserProfile) error {
    query := `
        UPDATE user_profiles
        SET preferences = $1,
            home_location_lat = $2,
            home_location_lon = $3,
            dietary_restrictions = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $5
        RETURNING updated_at`

    return s.db.QueryRow(
        query,
        profile.Preferences,
        profile.HomeLocationLat,
        profile.HomeLocationLon,
        pq.Array(profile.DietaryRestrictions),
        profile.ID,
    ).Scan(&profile.UpdatedAt)
}

func (s *userStore) GetNearbyUsers(lat, lon float64, radiusKm float64) ([]types.UserProfile, error) {
    query := `
        SELECT id, preferences, home_location_lat, home_location_lon,
               dietary_restrictions, created_at, updated_at
        FROM user_profiles
        WHERE point(home_location_lon, home_location_lat) <@> point($1, $2) <= $3
        ORDER BY point(home_location_lon, home_location_lat) <@> point($1, $2)`

    rows, err := s.db.Query(query, lon, lat, radiusKm)
    if err != nil {
        return nil, fmt.Errorf("error querying nearby users: %v", err)
    }
    defer rows.Close()

    var profiles []types.UserProfile
    for rows.Next() {
        var p types.UserProfile
        err := rows.Scan(
            &p.ID,
            &p.Preferences,
            &p.HomeLocationLat,
            &p.HomeLocationLon,
            pq.Array(&p.DietaryRestrictions),
            &p.CreatedAt,
            &p.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning user row: %v", err)
        }
        profiles = append(profiles, p)
    }
    return profiles, rows.Err()
}

func (s *userStore) CreateGroup(name string, userID int) (int, error) {
    var groupID int
    query := `
        INSERT INTO groups (name, created_by)
        VALUES ($1, $2)
        RETURNING id`

    err := s.db.QueryRow(query, name, userID).Scan(&groupID)
    return groupID, err
}

func (s *userStore) AddGroupMember(groupID int, userID int, searchRadiusMeters int) error {
    query := `
        INSERT INTO group_members (group_id, user_profile_id, search_radius_meters)
        VALUES ($1, $2, $3)`

    _, err := s.db.Exec(query, groupID, userID, searchRadiusMeters)
    return err
}

func (s *userStore) GetGroupMembers(groupID int) ([]types.UserProfile, error) {
    query := `
        SELECT p.id, p.preferences, p.home_location_lat, p.home_location_lon,
               p.dietary_restrictions, p.created_at, p.updated_at
        FROM user_profiles p
        JOIN group_members gm ON gm.user_profile_id = p.id
        WHERE gm.group_id = $1`

    rows, err := s.db.Query(query, groupID)
    if err != nil {
        return nil, fmt.Errorf("error querying group members: %v", err)
    }
    defer rows.Close()

    var profiles []types.UserProfile
    for rows.Next() {
        var p types.UserProfile
        err := rows.Scan(
            &p.ID,
            &p.Preferences,
            &p.HomeLocationLat,
            &p.HomeLocationLon,
            pq.Array(&p.DietaryRestrictions),
            &p.CreatedAt,
            &p.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning member row: %v", err)
        }
        profiles = append(profiles, p)
    }
    return profiles, rows.Err()
}
