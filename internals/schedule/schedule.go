package schedule

import (
	"fmt"
	"time"

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
