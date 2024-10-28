package components

import (
	"context"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func Render(ctx echo.Context, status int, t templ.Component) error {
    ctx.Response().Writer.WriteHeader(status)

    err := t.Render(context.Background(), ctx.Response().Writer)
    if err != nil {
        log.Println(err)
        return ctx.String(http.StatusInternalServerError, "failed to render response template")
    }

    return nil
}
