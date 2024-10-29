package main

import (
	"fmt"
	"time"

	"github.com/rickb777/date"
)

type Schedule struct {
	Name  string
	RDOs  RDO
	Start date.Date
}

type RDO struct {
	count int
	days  []time.Weekday
}

type Week [7]int32

type Set struct {
	Days      []date.Date
	Protected bool
	Available bool
}

func contains(s []time.Weekday, e time.Weekday) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func print_rdos(s date.Date, rdos []time.Weekday, r int) {
	set := Set{}
	for i := range r {
		d := s.Add(date.PeriodOfDays(i))
		status := contains(rdos, d.Weekday())
		if status {
			set.Days = append(set.Days, d)
		}
	}
	fmt.Println("Rdo set:")
	for _, v := range set.Days {
		fmt.Printf("%s - %s \n", v, v.Weekday())
	}
	return
}

func generate_set(s date.Date, rdos []time.Weekday, r int) Set {
	set := Set{}
	for i := range r {
		d := s.Add(date.PeriodOfDays(i))
		status := contains(rdos, d.Weekday())
		if status {
			set.Days = append(set.Days, d)
		}
	}
	return set
}

func generate_schedule(s Schedule, t date.Date) {
	count := t.Sub(s.Start)
	fmt.Println("Number of days between start & target:", count)
	weeks := count / 7
	remainder := count % 7
	fmt.Printf("%d weeks and %d days", weeks, remainder)
}

func main() {
	fmt.Println("Hello, let's start solving problems!")
	standard := Schedule{
		Name: "standard",
		RDOs: RDO{
			count: 2,
			days:  []time.Weekday{0, 6},
		},
		Start: date.Today(),
	}
	fmt.Printf("Controller A is on a %s schedule of %d rdos, %s and %s. \n", standard.Name, standard.RDOs.count, standard.RDOs.days[0], standard.RDOs.days[1])
	generate_schedule(standard, standard.Start.Add(31))
}
