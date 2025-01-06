package types

import (
	"encoding/json"
	"time"
)

// RestaurantChain represents a restaurant brand/chain
type RestaurantChain struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Website     string    `json:"website,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// RestaurantMetadata represents a specific restaurant location
type RestaurantMetadata struct {
	ItemID         int             `json:"item_id"`
	ChainID        *int            `json:"chain_id,omitempty"`
	PriceRange     int             `json:"price_range"`
	Latitude       float64         `json:"latitude"`
	Longitude      float64         `json:"longitude"`
	Address        string          `json:"address"`
	OperatingHours json.RawMessage `json:"operating_hours"`
	Website        string          `json:"website,omitempty"`
	Phone          string          `json:"phone,omitempty"`
	GooglePlaceID  string          `json:"google_place_id"`
    GoogleMapsUri  string          `json:"google_maps_uri"`
}

// GooglePlacesResponse represents the relevant fields from Google Places API
type GooglePlacesResponse struct {
    Name                     string          `json:"display_name"`
    Rating                   float64         `json:"rating"`
    Location                 struct {
        Latitude  float64 `json:"latitude"`
        Longitude float64 `json:"longitude"`
    } `json:"location"`
    FormattedAddress        string          `json:"formatted_address"`
    RegularOpeningHours     json.RawMessage `json:"regular_opening_hours"`
    Website                  string          `json:"website_uri"`
    InternationalPhoneNumber string          `json:"international_phone_number"`
    GoogleMapsUri           string          `json:"google_maps_uri"`
    PriceRange              struct {
        StartPrice struct {
            CurrencyCode string `json:"currency_code"`
            Units       int64  `json:"units"`
        } `json:"start_price"`
        EndPrice struct {
            CurrencyCode string `json:"currency_code"`
            Units       int64  `json:"units"`
        } `json:"end_price"`
    } `json:"price_range"`
}

// RestaurantStore defines the interface for restaurant data operations
type RestaurantStore interface {
    CreateChain(chain *RestaurantChain) error
    GetChainByID(id int) (*RestaurantChain, error)
    CreateRestaurant(item *Item, metadata *RestaurantMetadata) error
    GetLocationsByChainID(chainID int) ([]RestaurantMetadata, error)
    GetNearbyLocations(lat, lon float64, radiusKm float64) ([]RestaurantMetadata, error)
    UpdateLocationFromGooglePlaces(placeID string, data *GooglePlacesResponse) error
}
