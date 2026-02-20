package schedules

import (
	"errors"
	"time"
)

type ScheduleService struct{}

func NewScheduleService() *ScheduleService {
	return &ScheduleService{}
}

// ListScheduleItemByDate
func (s *ScheduleService) ListScheduleItemByDate(schedule *CycledSchedule, educationStartDate time.Time, date time.Time) ([]ScheduleItem, error) {
	if s == nil {
		return nil, errors.New("schedule can not be nil")
	}

	if date.Before(educationStartDate) {
		return nil, errors.New("invalid date")
	}

	days := int(date.Truncate(24*time.Hour).
		Sub(educationStartDate.Truncate(24*time.Hour)).
		Hours() / 24)

	weekNumber := days / 7 + 1

	weekType := WeekTypeUneven
	if weekNumber%2 == 0 {
		weekType = WeekTypeEven
	}

	weekday := date.Weekday()
	if weekday == time.Sunday {
		return nil, nil
	}

	var result []ScheduleItem

	dayItems := schedule.ListItemByWeekday(weekday)
	for _, item := range dayItems {
		if *item.Weektype == weekType || *item.Weektype == WeekTypeBoth {
			item.Date = &date
			item.Weeknum = &weekNumber
			result = append(result, item)
		}
	}

	return result, nil
}
