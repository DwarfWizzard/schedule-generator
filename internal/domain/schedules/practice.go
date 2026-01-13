package schedules

import (
	"errors"
	"time"
)

type PracticeType int8

const (
	PracticeTypeIndustrial = iota
	PracticeTypeDiploma
)

var practiceNames = []string{
	"industrial",
	"diploma",
}

func (p PracticeType) String() string {
	return practiceNames[p]
}

func NewPracticeType(pt int8) (PracticeType, error) {
	if int(pt) < 0 || int(pt) >= len(practiceNames) {
		return 0, errors.New("unknown week type")
	}

	return PracticeType(pt), nil
}

type Practice struct {
	Type      PracticeType
	StartDate time.Time
	EndDate   time.Time
}
