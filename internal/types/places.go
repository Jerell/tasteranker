package types

import (
	"encoding/json"
	"time"
)

// RestaurantChain represents a restaurant brand/chain
type RestaurantChain struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Website       string    `json:"website,omitempty"`
	Description   string    `json:"description,omitempty"`
	FoundedYear   int       `json:"founded_year,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// RestaurantMetadata represents a specific restaurant location
type RestaurantMetadata struct {
	ItemID        int             `json:"item_id"`
	ChainID       int             `json:"chain_id"`
	CuisineType   string          `json:"cuisine_type"`
	PriceRange    int             `json:"price_range"` // 1-4 matching Google's price level
	Latitude      float64         `json:"latitude"`
	Longitude     float64         `json:"longitude"`
	Address       string          `json:"address"`
	OperatingHours json.RawMessage `json:"operating_hours"` // Stores complex hours format
	Website       string          `json:"website,omitempty"`
	Phone         string          `json:"phone,omitempty"`
	GooglePlaceID string          `json:"google_place_id"`
	Rating        float64         `json:"rating"`
	UserRatingsCount int          `json:"user_ratings_count"`
}

// UserProfile represents a user's preferences and location
type UserProfile struct {
	UserID              string          `json:"user_id"`
	Preferences         json.RawMessage `json:"preferences"` // Stores cuisine preferences, dietary restrictions, etc.
	HomeLocationLat     float64         `json:"home_location_lat"`
	HomeLocationLon     float64         `json:"home_location_lon"`
	DietaryRestrictions []string        `json:"dietary_restrictions"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// GooglePlacesResponse represents the relevant fields from Google Places API
type GooglePlacesResponse struct {
	PlaceID     string   `json:"place_id"`
	Name        string   `json:"name"`
	Rating      float64  `json:"rating"`
	UserRatings int      `json:"user_ratings_total"`
	PriceLevel  int      `json:"price_level"` // 0-4, where 0 means no data
	Types       []string `json:"types"`
	Geometry    struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
	FormattedAddress string          `json:"formatted_address"`
	OpeningHours     json.RawMessage `json:"opening_hours"`
	Website          string          `json:"website"`
	InternationalPhoneNumber string  `json:"international_phone_number"`
}

// RestaurantStore defines the interface for restaurant data operations
type RestaurantStore interface {
	CreateChain(chain *RestaurantChain) error
	GetChainByID(id int) (*RestaurantChain, error)
	CreateLocation(metadata *RestaurantMetadata) error
	GetLocationsByChainID(chainID int) ([]RestaurantMetadata, error)
	GetNearbyLocations(lat, lon float64, radiusKm float64) ([]RestaurantMetadata, error)
	UpdateLocationFromGooglePlaces(placeID string, data *GooglePlacesResponse) error
}

// UserStore defines the interface for user data operations
type UserStore interface {
	CreateProfile(profile *UserProfile) error
	GetProfile(userID string) (*UserProfile, error)
	UpdateProfile(profile *UserProfile) error
	GetNearbyUsers(lat, lon float64, radiusKm float64) ([]UserProfile, error)
}
