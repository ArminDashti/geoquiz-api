package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/armin/geoquiz-api/internal/auth"
	"github.com/armin/geoquiz-api/internal/config"
	"github.com/armin/geoquiz-api/internal/data"
	appdb "github.com/armin/geoquiz-api/internal/db"
	"github.com/armin/geoquiz-api/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadDotEnv(".env")

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	store, err := data.LoadCountries(cfg.GeoJSONPath)
	if err != nil {
		log.Fatalf("load countries: %v", err)
	}

	sqlDB, err := appdb.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer sqlDB.Close()

	if err := appdb.Migrate(sqlDB, cfg.MigrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := appdb.SeedInviteCode(ctx, sqlDB, cfg.InviteCode); err != nil {
		log.Fatalf("seed invite code: %v", err)
	}
	if err := appdb.SeedBootstrapUser(ctx, sqlDB, "armin", "armin@geoquiz.local", "dopadopa123", "Armin", "Dashti"); err != nil {
		log.Fatalf("seed bootstrap user: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(cfg.UploadDir, "avatars"), 0o755); err != nil {
		log.Fatalf("upload dir: %v", err)
	}

	h := handlers.New(store, sqlDB, cfg)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/uploads", cfg.UploadDir)

	r.GET("/health", h.Health)
	api := r.Group("/api/v1")
	{
		api.GET("/countries", h.ListCountries)
		api.GET("/countries/geojson", h.CountriesGeoJSON)
		api.GET("/countries/:id/neighbors", h.CountryNeighbors)

		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)
		api.GET("/profiles/:username", h.GetProfile)
		api.GET("/scores", h.ListScores)

		authed := api.Group("")
		authed.Use(auth.Middleware(cfg.JWTSecret))
		{
			authed.GET("/auth/me", h.Me)
			authed.PATCH("/account", h.UpdateAccount)
			authed.POST("/account/avatar", h.UploadAvatar)
			authed.POST("/account/password", h.ChangePassword)
			authed.DELETE("/account", h.DeleteAccount)
			authed.POST("/scores", h.CreateScore)

			admin := authed.Group("/admin")
			admin.Use(auth.RequireAdmin())
			{
				admin.GET("/invite-code", h.GetInviteCode)
				admin.PUT("/invite-code", h.UpdateInviteCode)
			}
		}
	}

	log.Printf("geoquiz-api listening on %s (geojson=%s)", cfg.Addr, cfg.GeoJSONPath)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
