package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"

	csvHandlers "github.com/AshwinShukla72/csv-processor/handlers"
	"github.com/AshwinShukla72/csv-processor/repositories"
	"github.com/AshwinShukla72/csv-processor/services"
	"github.com/AshwinShukla72/csv-processor/utils"
)

func main() {
	// Create a new Fiber instance with custom configuration
	app := fiber.New(fiber.Config{
		JSONEncoder:   json.Marshal,   // Use standard JSON encoder
		JSONDecoder:   json.Unmarshal, // Use standard JSON decoder
		Prefork:       true,
		CaseSensitive: true, // Enabling case-sensitive routing
		StrictRouting: true, // Enabling StrictRouting
		ServerHeader:  "Fiber",
		AppName:       "CSV Processor Service",
	})

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // Optimizing for speed over compression ratio
	}))

	app.Use(idempotency.New())
	app.Use(helmet.New())
	app.Use(limiter.New())
	app.Use(etag.New())

	// Log incoming requests with custom format
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n", // Log format
	}))

	// Expose application metrics at /metrics (default config)
	app.Get("/metrics", monitor.New())

	// Overwrite /metrics with a custom title (this will override the previous route)
	app.Get("/metrics", monitor.New(monitor.Config{
		Title: "CSV Processor Metrics Page", // Custom title for metrics page
	}))

	// Add a health check endpoint at /health ->
	// Liveness Endpoint: /livez
	// Add a readiness endpoint at /readyz
	app.Use(healthcheck.New())

	store := repositories.NewStore("files")
	jobs := repositories.NewRedisQueue("localhost:6379")
	parser := services.NewCSVParser()

	// Start background worker to process jobs
	go jobs.Worker(store, parser)

	app.Post("/API/upload", csvHandlers.UploadHandler(store, jobs, parser))
	app.Get("/API/download/:id", csvHandlers.DownloadHandler(store, jobs))
	// Start the server on port 3000, No env at the moment
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start the Fiber app in a separate goroutine to avoid blocking the main thread.
	go func() {
		addr := ":" + port
		if err := app.Listen(addr); err != nil {
			// If the server fails to start, log the error and terminate the application
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Listen for termination signals (e.g., SIGINT, SIGTERM) and gracefully shut down the server.
	utils.GracefulShutdown(app, 5*time.Second)
}
