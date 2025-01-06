package types

import (
	"encoding/json"
	"time"
)

// UserProfile represents a user's preferences and location
type UserProfile struct {
    ID                  int             `json:"id"`
    Preferences         json.RawMessage `json:"preferences"`
    HomeLocationLat     float64         `json:"home_location_lat"`
    HomeLocationLon     float64         `json:"home_location_lon"`
    DietaryRestrictions []string        `json:"dietary_restrictions"`
    CreatedAt          time.Time       `json:"created_at"`
    UpdatedAt          time.Time       `json:"updated_at"`
}

type AuthProvider struct {
    ID             int       `json:"id"`
    Provider       string    `json:"provider"`
    ProviderUserID string    `json:"provider_user_id"`
    UserProfileID  int       `json:"user_profile_id"`
    CreatedAt      time.Time `json:"created_at"`
}

// UserStore defines the interface for user data operations
type UserStore interface {
    CreateProfile(profile *UserProfile) error
    GetProfile(id int) (*UserProfile, error)
    UpdateProfile(profile *UserProfile) error
    GetNearbyUsers(lat, lon float64, radiusKm float64) ([]UserProfile, error)
    GetProfileByProviderID(provider, providerUserID string) (*UserProfile, error)
    CreateAuthProvider(auth *AuthProvider) error

    // Group methods
    CreateGroup(name string, userID int) (int, error)
    AddGroupMember(groupID int, userID int, searchRadiusMeters int) error
    GetGroupMembers(groupID int) ([]UserProfile, error)
}
