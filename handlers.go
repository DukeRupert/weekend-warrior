package main

import (
	"time"
	"strconv"
	"github.com/gofiber/fiber/v2"
)
// HTTP handler would need to be updated to include the pairs:
func CalendarHandler(c *fiber.Ctx) error {

	// Handle url query values
    year, month := GetCurrentYearMonth() // Default to current date
    if yearStr := c.Query("year"); yearStr != "" {
        if y, err := strconv.Atoi(yearStr); err == nil {
            year = y
        }
    }
    
    if monthStr :=c.Query("month"); monthStr != "" {
        if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
            month = m
        }
    }

    // Generate weekday pairs (example using Monday and Wednesday)
    anchorDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    pairs := GenerateWeekdayPairs(time.Monday, time.Wednesday, anchorDate)
    cal := GenerateCalendar(year, month, pairs)
    return c.Render("calendar", cal)
}