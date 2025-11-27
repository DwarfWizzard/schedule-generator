package common

import "time"

var DefaultTimezone *time.Location

func init() {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}

	DefaultTimezone = loc
}
