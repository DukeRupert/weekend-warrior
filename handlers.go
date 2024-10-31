package main

import (
	"time"
	"strconv"
	"github.com/gofiber/fiber/v2"
)

type TemplateData struct {
    Calendars []Calendar
}

// HTTP handler would need to be updated to include the pairs:
func CalendarHandler(c *fiber.Ctx) error {
	// Handle url query values
    year, month := GetCurrentYearMonth()
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

    // Generate the calendar for a specific pair set
    // anchorDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    // pairs := GenerateWeekdayPairs(time.Monday, time.Wednesday, anchorDate)
    // cal := GenerateCalendar(year, month, pairs, "JD", 0) // JD = initials, 0 = color index

    // Generate multiple calendars
    var calendars []Calendar

    anchorDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    // Example: Generate calendars for different pairs
    pairs1 := GenerateWeekdayPairs(time.Monday, time.Wednesday, anchorDate)
    cal1 := GenerateCalendar(year, month, pairs1, "JD", 0)

    pairs2 := GenerateWeekdayPairs(time.Tuesday, time.Thursday, anchorDate)
    cal2 := GenerateCalendar(year, month, pairs2, "AB", 1)

    calendars = append(calendars, cal1, cal2)

    data := TemplateData{
        Calendars: calendars,
    }

    return c.Render("calendar", data)
}