package places

import (
    "context"
    "net/http"
    "strconv"
    "fmt"
    
    places "cloud.google.com/go/maps/places/apiv1"
    placespb "cloud.google.com/go/maps/places/apiv1/placespb"
    "github.com/labstack/echo/v4"
    "github.com/Jerell/tasteranker/internal/types"
    "google.golang.org/genproto/googleapis/type/latlng"
)

type Store interface {
    CreateChain(chain *types.RestaurantChain) error
    GetChainByID(id int) (*types.RestaurantChain, error)
    CreateLocation(metadata *types.RestaurantMetadata) error
    GetLocationsByChainID(chainID int) ([]types.RestaurantMetadata, error)
    GetNearbyLocations(lat, lon float64, radiusKm float64) ([]types.RestaurantMetadata, error)
    UpdateLocationFromGooglePlaces(placeID string, data *types.GooglePlacesResponse) error
}

type handler struct {
    store        Store
    placesClient *places.Client
}

func UseSubroute(group *echo.Group, store Store) error {
    ctx := context.Background()
    client, err := places.NewClient(ctx)
    if err != nil {
        return fmt.Errorf("failed to create places client: %v", err)
    }

    h := &handler{
        store:        store,
        placesClient: client,
    }

    group.POST("/import", h.handleGoogleMapsImport)
    group.GET("/nearby", h.handleGetNearby)
    
    chainGroup := group.Group("/chains")
    chainGroup.POST("", h.handleCreateChain)
    chainGroup.GET("/:id", h.handleGetChain)
    chainGroup.GET("/:id/locations", h.handleGetChainLocations)

    return nil
}

func (h *handler) handleGetNearby(c echo.Context) error {
    lat, err := strconv.ParseFloat(c.QueryParam("lat"), 64)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid latitude")
    }

    lon, err := strconv.ParseFloat(c.QueryParam("lon"), 64)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid longitude")
    }

    radius, err := strconv.ParseFloat(c.QueryParam("radius"), 64)
    if err != nil {
        radius = 5.0 // Default 5km radius
    }

    req := &placespb.SearchNearbyRequest{
        LocationRestriction: &placespb.SearchNearbyRequest_LocationRestriction{
            Type: &placespb.SearchNearbyRequest_LocationRestriction_Circle{
                Circle: &placespb.Circle{
                    Center: &latlng.LatLng{
                        Latitude:  lat,
                        Longitude: lon,
                    },
                    Radius: float64(radius * 1000), // Convert km to meters
                },
            },
        },
        MaxResultCount: 20,
        RankPreference: placespb.SearchNearbyRequest_POPULARITY,
        IncludedTypes: []string{"restaurant"},
    }

    resp, err := h.placesClient.SearchNearby(c.Request().Context(), req)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to search nearby places: %v", err))
    }

    places := make([]types.RestaurantMetadata, 0, len(resp.Places))
    for _, place := range resp.Places {
        if place.Location == nil {
            continue
        }
        
        metadata := types.RestaurantMetadata{
            ItemID:      0, // This will be set when saving to DB
            ChainID:     0, // This will be set when saving to DB
            CuisineType: "", // Would need to parse from place.Types
            PriceRange:  int(place.PriceLevel),
            Latitude:    place.Location.Latitude,
            Longitude:   place.Location.Longitude,
            Address:     place.FormattedAddress,
            Website:     place.WebsiteUri,
            Phone:       place.InternationalPhoneNumber,
            Rating:      float64(place.Rating),
            UserRatingsCount: int(*place.UserRatingCount),
        }
        places = append(places, metadata)
    }

    return c.JSON(http.StatusOK, places)
}

func (h *handler) handleCreateChain(c echo.Context) error {
    chain := new(types.RestaurantChain)
    if err := c.Bind(chain); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid chain data")
    }

    if err := h.store.CreateChain(chain); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create chain")
    }

    return c.JSON(http.StatusCreated, chain)
}

func (h *handler) handleGetChain(c echo.Context) error {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid chain ID")
    }

    chain, err := h.store.GetChainByID(id)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch chain")
    }
    if chain == nil {
        return echo.NewHTTPError(http.StatusNotFound, "Chain not found")
    }

    return c.JSON(http.StatusOK, chain)
}

func (h *handler) handleGetChainLocations(c echo.Context) error {
    chainID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid chain ID")
    }

    locations, err := h.store.GetLocationsByChainID(chainID)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch locations")
    }

    return c.JSON(http.StatusOK, locations)
}

func (h *handler) handleGoogleMapsImport(c echo.Context) error {
    // Implementation needed
    return echo.NewHTTPError(http.StatusNotImplemented, "Google Maps import not yet implemented")
}

func convertGooglePlace(place *placespb.Place) *types.GooglePlacesResponse {
    if place == nil {
        return nil
    }

    response := &types.GooglePlacesResponse{
        PlaceID:     place.Name[7:], // Remove "places/" prefix
        Name:        place.DisplayName.Text,
        Rating:      float64(place.Rating),
        UserRatings: int(*place.UserRatingCount),
        PriceLevel:  int(place.PriceLevel),
        FormattedAddress: place.FormattedAddress,
        Website:     place.WebsiteUri,
        InternationalPhoneNumber: place.InternationalPhoneNumber,
    }

    if place.Location != nil {
        response.Geometry.Location.Lat = place.Location.Latitude
        response.Geometry.Location.Lng = place.Location.Longitude
    }

    return response
}
