// handlers/facilities_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "testing"

    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/dukerupert/weekend-warrior/config"
    "github.com/dukerupert/weekend-warrior/db"
    "github.com/dukerupert/weekend-warrior/db/models"
)

// testApp holds the test application instance
type testApp struct {
    app        *fiber.App
    dbService  *db.Service
    handler    *FacilityHandler
}

// setupTestApp creates a new Fiber app and handler for testing
func setupTestApp(t *testing.T, cfg *config.Config) *testApp {
    t.Helper()

    // Initialize DB service with test database connection
    dbService, err := db.NewService(db.Config{
        URL: cfg.GetDatabaseURL()
    })
    if err != nil {
        return nil, fmt.Errorf("unable to initialize database service: %v", err)
    }
    
    require.NoError(t, err)

    // Create Fiber app
    app := fiber.New()

    // Create handler
    handler := NewFacilityHandler(dbService)
    handler.RegisterRoutes(app)

    return &testApp{
        app:       app,
        dbService: dbService,
        handler:   handler,
    }
}

// clearFacilities removes all facilities from the test database
func clearFacilities(t *testing.T, dbService *db.Service) {
    t.Helper()
    _, err := dbService.GetPool().Exec(
        context.Background(),
        "DELETE FROM facilities",
    )
    require.NoError(t, err)
}

func TestFacilityEndpoints(t *testing.T) {
    // Load configuration
    cfg, err := config.LoadConfig(".env")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    app := setupTestApp(t, cfg)
    defer app.dbService.Close()

    t.Run("CRUD Operations", func(t *testing.T) {
        // Clear the database before testing
        clearFacilities(t, app.dbService)

        // Test: Create Facility
        t.Run("Create Facility", func(t *testing.T) {
            payload := map[string]string{
                "name": "Main Building",
                "code": "MAIN",
            }
            jsonBytes, err := json.Marshal(payload)
            require.NoError(t, err)

            req := httptest.NewRequest(
                "POST",
                "/facilities",
                bytes.NewReader(jsonBytes),
            )
            req.Header.Set("Content-Type", "application/json")

            resp, err := app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

            // Verify response body
            var response struct {
                Data models.Facility `json:"data"`
            }
            err = json.NewDecoder(resp.Body).Decode(&response)
            require.NoError(t, err)
            assert.Equal(t, "Main Building", response.Data.Name)
            assert.Equal(t, "MAIN", response.Data.Code)
        })

        // Test: List Facilities
        t.Run("List Facilities", func(t *testing.T) {
            req := httptest.NewRequest("GET", "/facilities", nil)
            resp, err := app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusOK, resp.StatusCode)

            // Verify response body
            var response struct {
                Data []models.Facility `json:"data"`
            }
            err = json.NewDecoder(resp.Body).Decode(&response)
            require.NoError(t, err)
            assert.Len(t, response.Data, 1) // Should have one facility from previous test
            assert.Equal(t, "Main Building", response.Data[0].Name)
        })

        // Test: Delete Facility by ID
        t.Run("Delete Facility by ID", func(t *testing.T) {
            // First, get the ID of the created facility
            req := httptest.NewRequest("GET", "/facilities", nil)
            resp, err := app.app.Test(req)
            require.NoError(t, err)

            var listResponse struct {
                Data []models.Facility `json:"data"`
            }
            err = json.NewDecoder(resp.Body).Decode(&listResponse)
            require.NoError(t, err)
            require.NotEmpty(t, listResponse.Data)

            facilityID := listResponse.Data[0].ID

            // Now delete the facility
            req = httptest.NewRequest(
                "DELETE",
                fmt.Sprintf("/facilities/%d", facilityID),
                nil,
            )
            resp, err = app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
        })

        // Test: Create another facility for delete by code test
        t.Run("Create Facility for Delete Test", func(t *testing.T) {
            payload := map[string]string{
                "name": "Another Building",
                "code": "TEST",
            }
            jsonBytes, err := json.Marshal(payload)
            require.NoError(t, err)

            req := httptest.NewRequest(
                "POST",
                "/facilities",
                bytes.NewReader(jsonBytes),
            )
            req.Header.Set("Content-Type", "application/json")

            resp, err := app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
        })

        // Test: Delete Facility by Code
        t.Run("Delete Facility by Code", func(t *testing.T) {
            req := httptest.NewRequest(
                "DELETE",
                "/facilities/code/TEST",
                nil,
            )
            resp, err := app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
        })
    })

    // Error cases
    t.Run("Error Cases", func(t *testing.T) {
        // Test: Create Facility with Invalid Code
        t.Run("Create Facility - Invalid Code", func(t *testing.T) {
            payload := map[string]string{
                "name": "Invalid Building",
                "code": "TOO_LONG",
            }
            jsonBytes, err := json.Marshal(payload)
            require.NoError(t, err)

            req := httptest.NewRequest(
                "POST",
                "/facilities",
                bytes.NewReader(jsonBytes),
            )
            req.Header.Set("Content-Type", "application/json")

            resp, err := app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
        })

        // Test: Delete Non-existent Facility
        t.Run("Delete Non-existent Facility", func(t *testing.T) {
            req := httptest.NewRequest(
                "DELETE",
                "/facilities/9999",
                nil,
            )
            resp, err := app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
        })

        // Test: Delete with Invalid ID
        t.Run("Delete with Invalid ID", func(t *testing.T) {
            req := httptest.NewRequest(
                "DELETE",
                "/facilities/invalid",
                nil,
            )
            resp, err := app.app.Test(req)
            require.NoError(t, err)
            assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
        })
    })
}