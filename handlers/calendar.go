package handlers

import (
	"strconv"
	"time"

	"github.com/dukerupert/weekend-warrior/services/calendar"
	"github.com/gofiber/fiber/v2"
)

type CalendarHandler struct {
	calendarService *calendar.Service
}

func NewCalendarHandler(calendarService *calendar.Service) *CalendarHandler {
	return &CalendarHandler{
		calendarService: calendarService,
	}
}

type TemplateData struct {
	Calendars []calendar.Calendar
}

// Example handler showing how to access DB from context
func (h *CalendarHandler) CalendarHandler(c *fiber.Ctx) error {
	// You can access the DB connection from context if needed
	// db := c.Locals("db").(*pgx.Conn)

	// Use db for queries...

	// Handle url query values
	year, month := h.calendarService.GetCurrentYearMonth()
	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	// Generate multiple calendars
	var calendars []calendar.Calendar

	anchorDate1 := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	anchorDate2 := time.Date(year, time.Month(month), 7, 0, 0, 0, 0, time.UTC)
	// Example: Generate calendars for different pairs
	pairs1 := h.calendarService.GenerateWeekdayPairs(time.Saturday, time.Sunday, anchorDate1)
	cal1 := h.calendarService.GenerateCalendar(year, month, pairs1, "YC", 0)

	pairs2 := h.calendarService.GenerateWeekdayPairs(time.Sunday, time.Monday, anchorDate2)
	cal2 := h.calendarService.GenerateCalendar(year, month, pairs2, "AB", 1)

	calendars = append(calendars, cal1, cal2)

	data := TemplateData{
		Calendars: calendars,
	}

	return c.Render("calendar", data)
}

