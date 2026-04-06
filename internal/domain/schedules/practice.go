package schedules

import (
	"fmt"
	"time"
)

type PracticeType int8

const (
	PracticeTypeEducational = iota
	PracticeTypeIndustrial
	PracticeTypeDiploma
)

var practiceNames = []string{
	"educational",
	"industrial",
	"diploma",
}

func (p PracticeType) String() string {
	return practiceNames[p]
}

func NewPracticeType(pt int8) (PracticeType, error) {
	if int(pt) < 0 || int(pt) >= len(practiceNames) {
		return 0, fmt.Errorf("unknown week type %d", pt)
	}

	return PracticeType(pt), nil
}

type Practice struct {
	Type      PracticeType
	StartDate time.Time
	EndDate   time.Time
}
