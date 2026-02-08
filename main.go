package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"omnitranscripts/config"
	"omnitranscripts/handlers"
	"omnitranscripts/jobs"
	"omnitranscripts/lib"
	"omnitranscripts/mcp"
)

func main() {
	cfg := config.Load()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(cors.New())
	app.Use(logger.New())

	jobs.Initialize()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "OmniTranscripts API is running",
		})
	})

	// Mount MCP server if enabled
	if cfg.MCPEnabled {
		mcpHandler := mcp.NewHTTPHandler(cfg.APIKey)
		// Use All to handle all HTTP methods (GET, POST, DELETE) required by MCP
		app.All(cfg.MCPEndpoint, adaptor.HTTPHandler(mcpHandler))
		app.All(cfg.MCPEndpoint+"/*", adaptor.HTTPHandler(mcpHandler))
		log.Printf("MCP server enabled at %s", cfg.MCPEndpoint)
	}

	api := app.Group("/", lib.AuthMiddleware())
	api.Post("/transcribe", handlers.PostTranscribe)
	api.Get("/transcribe/:job_id", handlers.GetTranscribeJob)

	log.Printf("Starting server on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
