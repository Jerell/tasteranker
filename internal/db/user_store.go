package db

import (
    "database/sql"
    "fmt"
    "github.com/lib/pq"
    "github.com/Jerell/tasteranker/internal/types"
)

type PostgresUserStore struct {
    db *sql.DB
}

func NewUserStore(db *sql.DB) *PostgresUserStore {
    return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) CreateProfile(profile *types.UserProfile) error {
    query := `
        INSERT INTO user_profiles (
            user_id, preferences, home_location_lat,
            home_location_lon, dietary_restrictions
        ) VALUES ($1, $2, $3, $4, $5)
        RETURNING created_at, updated_at`

    // Convert dietary restrictions to pq.StringArray for storage
    restrictions := pq.StringArray(profile.DietaryRestrictions)

    err := s.db.QueryRow(
        query,
        profile.UserID,
        profile.Preferences,
        profile.HomeLocationLat,
        profile.HomeLocationLon,
        restrictions,
    ).Scan(&profile.CreatedAt, &profile.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to create user profile: %v", err)
    }
    return nil
}

func (s *PostgresUserStore) GetProfile(userID string) (*types.UserProfile, error) {
    query := `
        SELECT user_id, preferences, home_location_lat,
               home_location_lon, dietary_restrictions,
               created_at, updated_at
        FROM user_profiles
        WHERE user_id = $1`

    profile := &types.UserProfile{}
    var restrictions pq.StringArray

    err := s.db.QueryRow(query, userID).Scan(
        &profile.UserID,
        &profile.Preferences,
        &profile.HomeLocationLat,
        &profile.HomeLocationLon,
        &restrictions,
        &profile.CreatedAt,
        &profile.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get user profile: %v", err)
    }

    profile.DietaryRestrictions = []string(restrictions)
    return profile, nil
}

func (s *PostgresUserStore) UpdateProfile(profile *types.UserProfile) error {
    query := `
        UPDATE user_profiles
        SET preferences = $1,
            home_location_lat = $2,
            home_location_lon = $3,
            dietary_restrictions = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $5
        RETURNING updated_at`

    restrictions := pq.StringArray(profile.DietaryRestrictions)

    err := s.db.QueryRow(
        query,
        profile.Preferences,
        profile.HomeLocationLat,
        profile.HomeLocationLon,
        restrictions,
        profile.UserID,
    ).Scan(&profile.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to update user profile: %v", err)
    }
    return nil
}

func (s *PostgresUserStore) GetNearbyUsers(lat, lon float64, radiusKm float64) ([]types.UserProfile, error) {
    // Using the Haversine formula to calculate distance
    query := `
        SELECT user_id, preferences, home_location_lat,
               home_location_lon, dietary_restrictions,
               created_at, updated_at
        FROM user_profiles
        WHERE (
            6371 * acos(
                cos(radians($1)) *
                cos(radians(home_location_lat)) *
                cos(radians(home_location_lon) - radians($2)) +
                sin(radians($1)) *
                sin(radians(home_location_lat))
            )
        ) <= $3
        ORDER BY created_at DESC`

    rows, err := s.db.Query(query, lat, lon, radiusKm)
    if err != nil {
        return nil, fmt.Errorf("failed to query nearby users: %v", err)
    }
    defer rows.Close()

    var users []types.UserProfile
    for rows.Next() {
        var user types.UserProfile
        var restrictions pq.StringArray

        err := rows.Scan(
            &user.UserID,
            &user.Preferences,
            &user.HomeLocationLat,
            &user.HomeLocationLon,
            &restrictions,
            &user.CreatedAt,
            &user.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan user row: %v", err)
        }

        user.DietaryRestrictions = []string(restrictions)
        users = append(users, user)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating user rows: %v", err)
    }

    return users, nil
}

func (s *PostgresUserStore) CreateGroup(name string, userID int) (int, error) {
    query := `
        INSERT INTO groups (name, created_by, status)
        VALUES ($1, $2, 'planning')
        RETURNING id`
        
    var groupID int
    err := s.db.QueryRow(query, name, userID).Scan(&groupID)
    if err != nil {
        return 0, fmt.Errorf("failed to create group: %v", err)
    }

    // Add creator as first member
    memberQuery := `
        INSERT INTO group_members (group_id, user_id)
        VALUES ($1, $2)`
        
    _, err = s.db.Exec(memberQuery, groupID, userID)
    if err != nil {
        return 0, fmt.Errorf("failed to add creator to group: %v", err)
    }

    return groupID, nil
}

func (s *PostgresUserStore) AddGroupMember(groupID int, userID int, searchRadiusMeters int) error {
    query := `
        INSERT INTO group_members (group_id, user_id, search_radius_meters)
        VALUES ($1, $2, $3)`
        
    _, err := s.db.Exec(query, groupID, userID, searchRadiusMeters)
    if err != nil {
        return fmt.Errorf("failed to add group member: %v", err)
    }
    
    return nil
}

func (s *PostgresUserStore) GetGroupMembers(groupID int) ([]types.UserProfile, error) {
    query := `
        SELECT u.user_id, u.preferences, u.home_location_lat,
               u.home_location_lon, u.dietary_restrictions,
               u.created_at, u.updated_at
        FROM user_profiles u
        JOIN group_members gm ON u.user_id = gm.user_id
        WHERE gm.group_id = $1
        ORDER BY gm.joined_at DESC`

    rows, err := s.db.Query(query, groupID)
    if err != nil {
        return nil, fmt.Errorf("failed to query group members: %v", err)
    }
    defer rows.Close()

    var members []types.UserProfile
    for rows.Next() {
        var member types.UserProfile
        var restrictions pq.StringArray

        err := rows.Scan(
            &member.UserID,
            &member.Preferences,
            &member.HomeLocationLat,
            &member.HomeLocationLon,
            &restrictions,
            &member.CreatedAt,
            &member.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan member row: %v", err)
        }

        member.DietaryRestrictions = []string(restrictions)
        members = append(members, member)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating member rows: %v", err)
    }

    return members, nil
}
