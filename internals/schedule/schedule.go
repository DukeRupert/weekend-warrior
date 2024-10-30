package schedule

import (
	"fmt"
	"time"
	"html/template"
    "bytes"
    "net/http"
    "strconv"
	"github.com/rickb777/date"
)

type (
	Week [7]int32
	Days []time.Weekday
)

type Schedule struct {
	Name string
	Bid  Bid
}

type Bid struct {
	Count  int
	Days   Days
	Anchor date.Date
}

type RDO struct {
	Days      []date.Date
	Protected bool
	Available bool
}

type RDOs []RDO

func NewSchedule() (*Schedule, error) {
	s := &Schedule{
		Name: "standard",
		Bid: Bid{
			Count:  2,
			Days:   Days{0, 6},
			Anchor: date.Today(),
		},
	}

	return s, nil
}

func (s *Schedule) contains(e time.Weekday) bool {
	for _, a := range s.Bid.Days {
		if a == e {
			return true
		}
	}
	return false
}

func (s *Schedule) generate_rdos(r date.PeriodOfDays) RDOs {
	schedule := []RDO{}
	set := RDO{Days: []date.Date{}, Protected: false, Available: false}
	for i := range r {
		d := s.Bid.Anchor.Add(date.PeriodOfDays(i))
		status := s.contains(d.Weekday())
		if status {
			set.Days = append(set.Days, d)
		}
		if len(set.Days) == 2 {
			schedule = append(schedule, set)
			set = RDO{}
		}
	}
	return schedule
}

// Given a schedule and target date return all RDO sets
func (s *Schedule) Generate_schedule(t date.Date) []RDO {
	// Calculate the difference in days between start date and target date
	count := t.Sub(s.Bid.Anchor)
	fmt.Println("Number of days between start & target:", count)

	// Generate rdo sets
	schedule := s.generate_rdos(count)
	schedule = Protect_sets(schedule)
	return schedule
}

func Protect_sets(s []RDO) []RDO {
	for i, v := range s {
		if i%3 == 0 {
			v.Protected = true
		}
	}
	return s
}

func Print_schedule(s []RDO) {
	for i, v := range s {
		if i%3 == 0 {
			v.Protected = true
		}
		if v.Protected {
			fmt.Println("This set of Rdos is protected -", v.Protected)
			for _, val := range v.Days {
				fmt.Printf("%s \n", val)
			}
		}
	}
}

// CalendarDay represents a single day in the calendar
type CalendarDay struct {
    Day     int  // 1-31, or 0 for empty cells
    IsToday bool // true if this is today's date
}

// Calendar represents a complete month calendar structure
type Calendar struct {
    Year        int
    Month       int
    Days        [][]CalendarDay
    MonthName   string
    YearOptions []int  // Added field for year dropdown
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

// GenerateCalendar now includes year options
func GenerateCalendar(year, month int) Calendar {
    // Get the current date for comparing with today
    currentYear, currentMonth := GetCurrentYearMonth()
    currentDay := time.Now().Day()

    // Generate year options (5 years before and after current year)
    yearOptions := make([]int, 11)
    for i := range yearOptions {
        yearOptions[i] = currentYear - 5 + i
    }

    // Get first day of week (0-6) and total days in month
    firstDayWeekday := FirstDayOfMonth(year, month)
    totalDays := DaysInMonth(year, month)

    // Initialize the calendar
    cal := Calendar{
        Year:        year,
        Month:       month,
        MonthName:   GetMonthName(month),
        YearOptions: yearOptions,
        Days:        make([][]CalendarDay, 6), // Maximum 6 weeks possible
    }

    // Initialize each week
    for i := range cal.Days {
        cal.Days[i] = make([]CalendarDay, 7)
    }

    // Fill in the calendar
    day := 1
    for week := 0; week < 6; week++ {
        for weekday := 0; weekday < 7; weekday++ {
            if week == 0 && weekday < firstDayWeekday {
                cal.Days[week][weekday] = CalendarDay{Day: 0, IsToday: false}
            } else if day <= totalDays {
                isToday := year == currentYear && month == currentMonth && day == currentDay
                cal.Days[week][weekday] = CalendarDay{Day: day, IsToday: isToday}
                day++
            } else {
                cal.Days[week][weekday] = CalendarDay{Day: 0, IsToday: false}
            }
        }
    }

    return cal
}
