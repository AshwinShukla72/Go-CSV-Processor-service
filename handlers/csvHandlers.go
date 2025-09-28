package csvHandlers

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/AshwinShukla72/csv-processor/repositories"
	"github.com/AshwinShukla72/csv-processor/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UploadHandler handles file uploads, validates file type, stores the original file,
// and queues it for processing. Only .csv and .txt files are accepted.
func UploadHandler(store repositories.Store, queue repositories.JobQueue, parser services.Parser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse a single file from the multipart form under key "file"
		fileHeader, err := c.FormFile("file")
		if err != nil {
			// missing file or bad multipart
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "missing file",
			})
		}

		// Validate extension
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext != ".csv" && ext != ".txt" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid file type",
			})
		}

		// Open the uploaded file
		f, err := fileHeader.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to read uploaded file",
			})
		}
		defer f.Close()

		// Read file into memory (keeps current behaviour). Consider streaming for larger files.
		data, err := io.ReadAll(f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to read uploaded file",
			})
		}

		// Generate job ID and save the original file
		id := uuid.New().String()
		if err := store.SaveOriginal(id, data); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save file",
			})
		}

		// Queue the job for processing (worker should pick this up)
		queue.Queue(id)

		// Respond with the generated job ID
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"id": id,
		})
	}
}

// DownloadHandler returns a Fiber handler that serves processed CSV data for a given job ID.
// It validates the UUID, checks job state, and returns the CSV if processing is complete.
func DownloadHandler(store repositories.Store, queue repositories.JobQueue) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ensure the client accepts JSON responses
		c.Accepts("application/json")

		// Extract job ID from route parameters
		rawID := c.Params("id")
		if rawID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
		}

		// Validate UUID format
		if _, err := uuid.Parse(rawID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}

		// Check job state; if still processing, return 423 Locked per previous behavior
		state, err := queue.State(rawID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job state"})
		}
		if state == "processing" {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{"error": "processing"})
		}

		// Load the processed CSV data from the store
		data, err := store.LoadProcessed(rawID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Not found",
			})
		}

		// Set response content type to CSV and send the data
		c.Type("csv")
		// Optionally suggest a filename:
		// c.Response().Header.Set("Content-Disposition", "attachment; filename=\"processed.csv\"")
		return c.Status(fiber.StatusOK).Send(data)
	}
}
