package main

import (
	"context"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Jerell/tasteranker/api/htmlcontent"
	"github.com/Jerell/tasteranker/api/users"
	"github.com/Jerell/tasteranker/components"
	"github.com/Jerell/tasteranker/internal/db"
	"github.com/Jerell/tasteranker/tigris"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/joho/godotenv"
)

func main() {
	e := echo.New()

	err := godotenv.Load()
	if err != nil {
		e.Logger.Warn("Error loading .env file in development")
	}

	dbConfig := db.NewConfig()
	database, err := db.NewConnection(dbConfig)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer database.Close()

	env := os.Getenv("APP_ENV")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println(env, port)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if env == "development" {
		e.Static("/assets", "./assets")
	} else {
		e.GET("/assets/*", func(c echo.Context) error {
			key := c.Param("*")
			ctx := context.Background()

			client, err := tigris.Client(ctx)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Failed to create Tigris client")
			}

			// Use the Tigris client to get the object
			resp, err := client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String("frosty-sound-5710"),
				Key:    aws.String("assets/" + key),
			})
			if err != nil {
				return c.String(http.StatusNotFound, "File not found")
			}
			defer resp.Body.Close()

			contentType := mime.TypeByExtension(filepath.Ext(key))
			if contentType == "" {
				contentType = "application/octet-stream" // Default fallback
			}

			// Stream the S3 object to the client
			return c.Stream(http.StatusOK, contentType, resp.Body)
		})
	}

	e.GET("/about", func(c echo.Context) error {
		return components.Render(
			c, http.StatusOK,
			components.Main(components.About()),
		)
	})

	e.GET("/", func(c echo.Context) error {
		return components.Render(
			c, http.StatusOK,
			components.Main(components.Home()),
		)
	})

	usersGroup := e.Group("/users/")
	userStore := db.NewUserStore(database)
	users.UseSubroute(usersGroup, userStore)

	htmlGroup := e.Group("/html/")
	htmlcontent.UseSubroute(htmlGroup)

	e.Logger.Fatal(e.Start(":" + port))
}
