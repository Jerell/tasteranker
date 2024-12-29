package auth

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var Store *sessions.CookieStore

func Init(baseURL string) {
    key := os.Getenv("SESSION_SECRET")
    if key == "" {
        panic("SESSION_SECRET environment variable is required")
    }

    Store = sessions.NewCookieStore([]byte(key))
    Store.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   86400 * 30,
        HttpOnly: true,
        Secure:   os.Getenv("APP_ENV") != "development",
        SameSite: http.SameSiteLaxMode,
    }

    gothic.Store = Store

    goth.UseProviders(
        google.New(
            os.Getenv("GOOGLE_CLIENT_ID"),
            os.Getenv("GOOGLE_CLIENT_SECRET"),
            baseURL+"/auth/google/callback",
            "email", "profile",
        ),
    )
}
