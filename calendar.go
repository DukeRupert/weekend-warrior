package main

import (
    "fmt"
	"time"
    "html/template"
)

// CalendarDay represents a single day in the calendar
type CalendarDay struct {
    Day       int
    IsToday   bool
    HasPair   bool
    Protected bool
}

// Calendar represents a complete month calendar structure
type Calendar struct {
    Year      int
    Month     int
    Days      [][]CalendarDay
    MonthName string
    Color     template.CSS    // HSL color for this calendar's pairs
    Initials  string    // Two-letter initials for the legend
}

// generateColor creates a pleasing HSL color based on an index
func generateColor(index int) template.CSS {
    // Use golden ratio for even color distribution
    goldenRatio := 0.618033988749895
    hue := float64(index) * goldenRatio
    
    // Keep the hue within [0,1)
    hue = hue - float64(int(hue))
    
    // Convert to degrees and create HSL color
    // Use 65% saturation and 60% lightness for pleasant, visible colors
    return template.CSS(fmt.Sprintf("hsl(%.0f, 65%%, 60%%)", hue*360))
}

// WeekdayPair represents a pair of weekdays
type WeekdayPair struct {
    First     time.Time
    Second    time.Time
    Protected bool
}

// GenerateWeekdayPairs generates pairs of weekdays for a year from the anchor date
func GenerateWeekdayPairs(firstWeekday, secondWeekday time.Weekday, anchorDate time.Time) []WeekdayPair {
    var pairs []WeekdayPair
    
    // Normalize time to midnight to ensure consistent date handling
    anchorDate = time.Date(
        anchorDate.Year(), 
        anchorDate.Month(), 
        anchorDate.Day(), 
        0, 0, 0, 0, 
        anchorDate.Location(),
    )
    
    // Find the first occurrence of the first weekday after or on the anchor date
    daysUntilFirst := (int(firstWeekday) - int(anchorDate.Weekday()) + 7) % 7
    if daysUntilFirst == 0 && anchorDate.Weekday() != firstWeekday {
        daysUntilFirst = 7
    }
    currentFirst := anchorDate.AddDate(0, 0, daysUntilFirst)
    
    // Calculate end date (1 year from anchor)
    endDate := anchorDate.AddDate(1, 0, 0)
    
    pairCount := 0  // Counter to track every third pair
    
    for currentFirst.Before(endDate) {
        // Find the next occurrence of the second weekday after the first weekday
        daysUntilSecond := (int(secondWeekday) - int(currentFirst.Weekday()) + 7) % 7
        if daysUntilSecond == 0 {
            daysUntilSecond = 7
        }
        currentSecond := currentFirst.AddDate(0, 0, daysUntilSecond)
        
        // Every third pair is protected
        isProtected := pairCount%3 == 0
        
        pairs = append(pairs, WeekdayPair{
            First:     currentFirst,
            Second:    currentSecond,
            Protected: isProtected,
        })
        
        // Move to next week and increment counter
        currentFirst = currentFirst.AddDate(0, 0, 7)
        pairCount++
    }
    
    return pairs
}

// GetMonthName converts month number to name
func GetMonthName(month int) string {
    months := []string{
        "January", "February", "March", "April",
        "May", "June", "July", "August",
        "September", "October", "November", "December",
    }
    if month < 1 || month > 12 {
        return ""
    }
    return months[month-1]
}

// GetCurrentYearMonth returns the current year and month as integers.
// Returns:
//   - year: The current year (e.g., 2024)
//   - month: The current month (1-12)
func GetCurrentYearMonth() (year int, month int) {
    now := time.Now()
    return now.Year(), int(now.Month())
}

// DaysInMonth returns the number of days in the specified year and month.
// year: The year (e.g., 2024)
// month: The month (1-12)
// Returns: The number of days in the specified month
// Example: DaysInMonth(2024, 2) returns 29 (leap year February)
func DaysInMonth(year int, month int) int {
    // Validate month input
    if month < 1 || month > 12 {
        return 0
    }
    
    // Create a time.Time for the first day of the next month
    firstOfNextMonth := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
    
    // Subtract one day to get the last day of our target month
    lastDay := firstOfNextMonth.AddDate(0, 0, -1)
    
    // Return the day component, which will be the number of days in the month
    return lastDay.Day()
}

// FirstDayOfMonth returns the weekday (0-6) of the first day of the specified month.
// Parameters:
//   - year: The year (e.g., 2024)
//   - month: The month (1-12)
// Returns:
//   - weekday: Integer from 0 (Sunday) to 6 (Saturday)
//   - If month is invalid, returns -1
func FirstDayOfMonth(year int, month int) int {
    // Validate month input
    if month < 1 || month > 12 {
        return -1
    }
    
    // Create a time.Time for the first day of the month
    firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    
    // Convert time.Weekday to int (Sunday = 0, Saturday = 6)
    return int(firstDay.Weekday())
}

// GenerateCalendar creates a calendar structure for a specific pair set
func GenerateCalendar(year, month int, pairs []WeekdayPair, initials string, colorIndex int) Calendar {
    // Get the current date for comparing with today
    currentYear, currentMonth := GetCurrentYearMonth()
    currentDay := time.Now().Day()

    firstDayWeekday := FirstDayOfMonth(year, month)
    totalDays := DaysInMonth(year, month)

    // Initialize the calendar
    cal := Calendar{
        Year:      year,
        Month:     month,
        MonthName: GetMonthName(month),
        Days:      make([][]CalendarDay, 6),
        Color:     generateColor(colorIndex),
        Initials:  initials,
    }

    // Create a map of dates to their pair status
    pairMap := make(map[time.Time]bool)
    protectedMap := make(map[time.Time]bool)
    
    for _, pair := range pairs {
        pairMap[pair.First] = true
        pairMap[pair.Second] = true
        if pair.Protected {
            protectedMap[pair.First] = true
            protectedMap[pair.Second] = true
        }
    }

    // Initialize and fill the calendar
    day := 1
    for i := range cal.Days {
        cal.Days[i] = make([]CalendarDay, 7)
        for j := range cal.Days[i] {
            if i == 0 && j < firstDayWeekday || day > totalDays {
                cal.Days[i][j] = CalendarDay{Day: 0}
                continue
            }

            currentDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
            cal.Days[i][j] = CalendarDay{
                Day:       day,
                IsToday:  year == currentYear && month == currentMonth && day == currentDay,
                HasPair:  pairMap[currentDate],
                Protected: protectedMap[currentDate],
            }
            day++
        }
    }

    return cal
}