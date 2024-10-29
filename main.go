package main

import (
	"fmt"
	"time"

	"github.com/rickb777/date"
)

type Schedule struct {
	Name  string
	Rdos  RDO
	Start date.Date
}

type RDO struct {
	count int
	days  []time.Weekday
}

func main() {
	fmt.Println("Hello, let's start solving problems!")
	standard := Schedule{
		Name: "standard",
		Rdos: RDO{
			count: 2,
			days:  []time.Weekday{0, 6},
		},
	}
	fmt.Printf("Controller A is on a %s schedule of %d rdos", standard.Name, standard.Rdos.count)
}
