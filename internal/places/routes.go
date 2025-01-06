package places

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	places "cloud.google.com/go/maps/places/apiv1"
	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/Jerell/tasteranker/internal/types"
	"github.com/labstack/echo/v4"
	"google.golang.org/genproto/googleapis/type/latlng"
	"google.golang.org/grpc/metadata"
)

type handler struct {
	store        types.RestaurantStore
	placesClient *places.Client
}

func UseSubroute(group *echo.Group, store types.RestaurantStore) error {
	ctx := context.Background()
	client, err := places.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create places client: %v", err)
	} else {
		println("created Places client")
	}

	h := &handler{
		store:        store,
		placesClient: client,
	}

	group.POST("/import", h.handleRestaurantImport)
	group.GET("/nearby", h.handleGetNearby)
	group.GET("/debug", h.handleDebugPlace)

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
		IncludedTypes:  []string{"restaurant"},
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
			ItemID:     0,
			PriceRange: int(place.PriceLevel),
			Latitude:   place.Location.Latitude,
			Longitude:  place.Location.Longitude,
			Address:    place.FormattedAddress,
			Website:    place.WebsiteUri,
			Phone:      place.InternationalPhoneNumber,
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

func getPriceLevel(priceRange *placespb.PriceRange) int {
	if priceRange == nil || priceRange.StartPrice == nil {
		return 1 // Default to lowest price level if not specified
	}

	// Using start price as the indicator
	startPrice := priceRange.StartPrice.Units
	endPrice := priceRange.EndPrice.Units

	// Average of start and end price
	avgPrice := (startPrice + endPrice) / 2

	// Map approximate price ranges to levels 1-4
	switch {
	case avgPrice <= 10:
		return 1
	case avgPrice <= 25:
		return 2
	case avgPrice <= 50:
		return 3
	default:
		return 4
	}
}

func (h *handler) handleRestaurantImport(c echo.Context) error {
    type importRequest struct {
        PlaceID string `json:"place_id"`
    }

    req := new(importRequest)
    if err := c.Bind(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
    }
    placeID := req.PlaceID

    ctx := context.Background()
    header := metadata.New(map[string]string{
        "X-Goog-FieldMask": "displayName,types,formattedAddress,regularOpeningHours,priceRange,location,internationalPhoneNumber,websiteUri,googleMapsUri",
    })
    ctx = metadata.NewOutgoingContext(ctx, header)

    place, err := h.placesClient.GetPlace(ctx, &placespb.GetPlaceRequest{
        Name: "places/" + placeID,
    })
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch place: %+v", err))
    }
    fmt.Printf("%+v\n", place)

    if !slices.Contains(place.Types, "restaurant") {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Not a restaurant: %+v", place.Types))
    }

    item := &types.Item{
        TypeID: 1,
        Name:   place.DisplayName.Text,
    }

    hoursJSON, err := json.Marshal(place.RegularOpeningHours)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("Failed to process operating hours: %v", err))
    }

    metadata := &types.RestaurantMetadata{
        PriceRange:     getPriceLevel(place.PriceRange),
        Latitude:       place.Location.Latitude,
        Longitude:      place.Location.Longitude,
        Address:        place.FormattedAddress,
        OperatingHours: json.RawMessage(hoursJSON),
        Website:        place.WebsiteUri,
        Phone:          place.InternationalPhoneNumber,
        GooglePlaceID:  placeID,
        GoogleMapsUri:  place.GoogleMapsUri,
    }

    fmt.Printf("%+v\n", metadata)

    if err := h.store.CreateRestaurant(item, metadata); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("Failed to create restaurant: %v", err))
    }

    return c.JSON(http.StatusCreated, metadata)
}

func (h *handler) handleDebugPlace(c echo.Context) error {
	placeID := c.QueryParam("place_id")
	if placeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "place_id query parameter required")
	}

	// Get the desired field mask from query parameter, default to all fields
	fieldMask := c.QueryParam("fields")
	if fieldMask == "" {
		fieldMask = "*" // Get all fields
	}

	// Create context with field mask header
	ctx := context.Background()
	header := metadata.New(map[string]string{
		"X-Goog-FieldMask": fieldMask,
	})
	ctx = metadata.NewOutgoingContext(ctx, header)

	place, err := h.placesClient.GetPlace(ctx, &placespb.GetPlaceRequest{
		Name: "places/" + placeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch place: %v", err))
	}

	return c.JSON(http.StatusOK, place)
}

func (h *handler) handleGoogleMapsImport(c echo.Context) error {
	type importRequest struct {
		PlaceID string `json:"place_id"`
		ChainID *int   `json:"chain_id,omitempty"` // Optional chain association
	}

	req := new(importRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	place, err := h.placesClient.GetPlace(c.Request().Context(), &placespb.GetPlaceRequest{
		Name: "places/" + req.PlaceID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch place details")
	}

	item := &types.Item{
		TypeID: 1,
		Name:   place.DisplayName.Text,
	}

	metadata := &types.RestaurantMetadata{
		ItemID:        item.ID,
		ChainID:       req.ChainID,
		PriceRange:    int(place.PriceLevel),
		Latitude:      place.Location.Latitude,
		Longitude:     place.Location.Longitude,
		Address:       place.FormattedAddress,
		Website:       place.WebsiteUri,
		Phone:         place.InternationalPhoneNumber,
		GooglePlaceID: req.PlaceID,
	}

	// Transaction to create both item and metadata
	if err := h.store.CreateRestaurant(item, metadata); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create restaurant")
	}

	return c.JSON(http.StatusCreated, metadata)
}


