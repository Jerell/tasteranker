package users

import (
    "net/http"
    "strconv"
    "github.com/labstack/echo/v4"
    "github.com/Jerell/tasteranker/internal/types"
)

type Store interface {
    CreateProfile(profile *types.UserProfile) error
    GetProfile(userID string) (*types.UserProfile, error)
    UpdateProfile(profile *types.UserProfile) error
    GetNearbyUsers(lat, lon float64, radiusKm float64) ([]types.UserProfile, error)
    CreateGroup(name string, userID int) (int, error)
    AddGroupMember(groupID int, userID int, searchRadiusMeters int) error
    GetGroupMembers(groupID int) ([]types.UserProfile, error)
}

type handler struct {
    store Store
}

func UseSubroute(group *echo.Group, store Store) {
    h := &handler{store: store}

    // User profile routes
    group.POST("/profiles", h.handleCreateProfile)
    group.GET("/profiles/:id", h.handleGetProfile)
    group.PUT("/profiles/:id", h.handleUpdateProfile)
    group.GET("/nearby", h.handleGetNearbyUsers)

    // Group routes
    groupRoutes := group.Group("/groups")
    groupRoutes.POST("", h.handleCreateGroup)
    groupRoutes.POST("/:id/members", h.handleAddGroupMember)
    groupRoutes.GET("/:id/members", h.handleGetGroupMembers)
}

func (h *handler) handleCreateProfile(c echo.Context) error {
    profile := new(types.UserProfile)
    if err := c.Bind(profile); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid profile data")
    }

    // Validate required fields
    if profile.UserID == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
    }

    if err := h.store.CreateProfile(profile); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to create profile: " + err.Error())
    }

    return c.JSON(http.StatusCreated, profile)
}

func (h *handler) handleGetProfile(c echo.Context) error {
    userID := c.Param("id")
    if userID == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
    }

    profile, err := h.store.GetProfile(userID)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to fetch profile: " + err.Error())
    }
    if profile == nil {
        return echo.NewHTTPError(http.StatusNotFound, "Profile not found")
    }

    return c.JSON(http.StatusOK, profile)
}

func (h *handler) handleUpdateProfile(c echo.Context) error {
    userID := c.Param("id")
    if userID == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
    }

    profile := new(types.UserProfile)
    if err := c.Bind(profile); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid profile data")
    }

    // Ensure URL parameter matches body
    if profile.UserID != userID {
        return echo.NewHTTPError(http.StatusBadRequest, "User ID mismatch")
    }

    if err := h.store.UpdateProfile(profile); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to update profile: " + err.Error())
    }

    return c.JSON(http.StatusOK, profile)
}

func (h *handler) handleGetNearbyUsers(c echo.Context) error {
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

    users, err := h.store.GetNearbyUsers(lat, lon, radius)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to fetch nearby users: " + err.Error())
    }

    return c.JSON(http.StatusOK, users)
}

func (h *handler) handleCreateGroup(c echo.Context) error {
    type createGroupRequest struct {
        Name   string `json:"name"`
        UserID int    `json:"user_id"`
    }

    req := new(createGroupRequest)
    if err := c.Bind(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request data")
    }

    if req.Name == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "Group name is required")
    }

    groupID, err := h.store.CreateGroup(req.Name, req.UserID)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to create group: " + err.Error())
    }

    return c.JSON(http.StatusCreated, map[string]int{"group_id": groupID})
}

func (h *handler) handleAddGroupMember(c echo.Context) error {
    groupID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid group ID")
    }

    type addMemberRequest struct {
        UserID            int `json:"user_id"`
        SearchRadiusMeters int `json:"search_radius_meters"`
    }

    req := new(addMemberRequest)
    if err := c.Bind(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request data")
    }

    if req.SearchRadiusMeters <= 0 {
        req.SearchRadiusMeters = 5000 // Default to 5km in meters
    }

    if err := h.store.AddGroupMember(groupID, req.UserID, req.SearchRadiusMeters); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to add group member: " + err.Error())
    }

    return c.NoContent(http.StatusCreated)
}

func (h *handler) handleGetGroupMembers(c echo.Context) error {
    groupID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid group ID")
    }

    members, err := h.store.GetGroupMembers(groupID)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to fetch group members: " + err.Error())
    }

    return c.JSON(http.StatusOK, members)
}
