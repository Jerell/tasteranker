package handlers

import (
	"net/http"

	"github.com/Jerell/tasteranker/internal/db"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
    store *db.UserStore
}

func NewUserHandler(store *db.UserStore) *UserHandler {
    return &UserHandler{store: store}
}

func (h *UserHandler) List(c echo.Context) error {
    limit := 50  // you could get this from query params
    offset := 0  // you could get this from query params
    
    users, err := h.store.List(c.Request().Context(), limit, offset)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": err.Error(),
        })
    }
    return c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c echo.Context) error {
    var input struct {
        Email string `json:"email"`
        Name  string `json:"name"`
    }
    
    if err := c.Bind(&input); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    user, err := h.store.Create(c.Request().Context(), input.Email, input.Name)
    if err != nil {
        switch err {
        case db.ErrDuplicateEmail:
            return c.JSON(http.StatusConflict, map[string]string{
                "error": "Email already exists",
            })
        case db.ErrInvalidUserData:
            return c.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid user data",
            })
        default:
            return c.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Internal server error",
            })
        }
    }
    
    return c.JSON(http.StatusCreated, user)
}
