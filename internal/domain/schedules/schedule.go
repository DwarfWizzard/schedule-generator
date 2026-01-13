package schedules

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidData  = errors.New("invalid data")
	ErrItemNotFound = errors.New("item not found")
	ErrItemConflict = errors.New("item conflict")
)

type ScheduleType int8

const (
	ScheduleTypeCycled ScheduleType = iota + 1
	ScheduleTypeCalendar
)

var scheduleTypeNames = []string{
	"cycled",
	"calendar",
}

func (w ScheduleType) String() string {
	return scheduleTypeNames[w]
}

type Schedule struct {
	ID         uuid.UUID
	EduGroupID uuid.UUID
	Semester   int
	Type       ScheduleType
	Cycled     *CycledSchedule
	Calendar   *CalendarSchedule
}

func (s *Schedule) ListItem() []ScheduleItem {
	if s == nil {
		return nil
	}

	switch s.Type {
	case ScheduleTypeCycled:
		return s.Cycled.ListItem()
	case ScheduleTypeCalendar:
		return s.Calendar.ListItem()
	}

	return nil
}

type CycledSchedule struct {
	StartDate time.Time
	EndDate   time.Time
	Items     map[time.Weekday][]ScheduleItem
}

// NewCycledSchedule
func NewCycledSchedule(eduGroupID uuid.UUID, semester int, startDate, endDate time.Time) (*Schedule, error) {
	id := uuid.New()

	if semester < 0 {
		return nil, errors.Join(ErrInvalidData, errors.New("invalid semester value"))
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.Join(ErrInvalidData, errors.New("both start and end dates can not be empty"))
	}

	if startDate.After(endDate) {
		return nil, errors.Join(ErrInvalidData, errors.New("start date is after end date"))
	}

	return &Schedule{
		ID:         id,
		EduGroupID: eduGroupID,
		Semester:   semester,
		Type:       ScheduleTypeCycled,
		Cycled: &CycledSchedule{
			StartDate: startDate,
			EndDate:   endDate,
			Items:     make(map[time.Weekday][]ScheduleItem, 6),
		},
	}, nil
}

// ListItem returns ScheduleItem array in weekday ordering
func (s CycledSchedule) ListItem() []ScheduleItem {
	var items []ScheduleItem
	for i := time.Monday; i <= time.Saturday; i++ {
		items = append(items, s.ListItemByWeekday(i)...)
	}

	return items
}

// ListItemByWeekday returns item list for specified weekday
func (s CycledSchedule) ListItemByWeekday(day time.Weekday) []ScheduleItem {
	return s.Items[day]
}

// AddItem adds item to schedule
func (s *CycledSchedule) AddItem(
	discipline string,
	teacherID uuid.UUID,
	weekday time.Weekday,
	studentsCount int16,
	lessonNumber int8,
	subgroup int8,
	weektype int8,
	lessonType int8,
	classroom string,
) error {
	var argErr error

	if lessonNumber < 0 {
		argErr = errors.Join(argErr, errors.New("invalid lesson number"))
	}

	if subgroup < 0 {
		argErr = errors.Join(argErr, errors.New("invalid subgroup"))
	}

	if weekday == time.Sunday {
		argErr = errors.Join(argErr, errors.New("item can not be created for sunday"))
	}

	if studentsCount < 0 {
		argErr = errors.Join(argErr, errors.New("invalid students count"))
	}

	if len(discipline) == 0 {
		argErr = errors.Join(argErr, errors.New("empty discipline"))
	}

	if len(classroom) == 0 {
		argErr = errors.Join(argErr, errors.New("empty classroom"))
	}

	wt, err := NewWeekType(weektype)
	if err != nil {
		argErr = errors.Join(argErr, err)
	}

	lt, err := NewItemLessonType(lessonType)
	if err != nil {
		argErr = errors.Join(argErr, err)
	}

	cr, err := NewClassroom(classroom)
	if err != nil {
		argErr = errors.Join(argErr, err)
	}

	if argErr != nil {
		return errors.Join(ErrInvalidData, argErr)
	}

	item := ScheduleItem{
		Discipline:    discipline,
		TeacherID:     teacherID,
		StudentsCount: studentsCount,
		Weekday:       weekday,
		LessonNumber:  lessonNumber,
		Subgroup:      subgroup,
		Weektype:      &wt,
		LessonType:    lt,
		Classroom:     cr,
	}

	if vErr := s.validateItem(&item); vErr != nil {
		return errors.Join(ErrInvalidData, vErr)
	}

	s.Items[item.Weekday] = append(s.Items[item.Weekday], item)

	return nil
}

// RemoveItem
func (s *CycledSchedule) RemoveItem(weekday time.Weekday, lessonNumber, subgroup, weektype int8) error {
	var argErr error

	if lessonNumber < 0 {
		argErr = errors.Join(argErr, errors.New("invalid lesson number"))
	}

	if subgroup < 0 {
		argErr = errors.Join(argErr, errors.New("invalid subgroup"))
	}

	wt, err := NewWeekType(weektype)
	if err != nil {
		argErr = errors.Join(argErr, err)
	}

	if argErr != nil {
		return errors.Join(ErrInvalidData, argErr)
	}

	idx := slices.IndexFunc(s.Items[weekday], func(item ScheduleItem) bool {
		return item.LessonNumber == lessonNumber && item.Subgroup == subgroup && *item.Weektype == wt
	})

	if idx < 0 {
		return ErrItemNotFound
	}

	s.Items[weekday] = append(s.Items[weekday][:idx], s.Items[weekday][idx+1:]...)

	return nil
}

func (s *CycledSchedule) validateItem(item *ScheduleItem) error {
	items := s.Items[item.Weekday]
	if len(items) == 0 {
		return nil
	}

	if item.Weektype == nil {
		return fmt.Errorf("weektype can not be empty in cycled schedule")
	}

	for _, current := range items {
		subgroupConflict :=
			current.Subgroup == item.Subgroup ||
				current.Subgroup == 0 ||
				item.Subgroup == 0

		conflict := (*current.Weektype == *item.Weektype && subgroupConflict) ||
			(*current.Weektype == WeekTypeBoth && (*item.Weektype != WeekTypeBoth || subgroupConflict)) ||
			(*item.Weektype == WeekTypeBoth && (*current.Weektype != WeekTypeBoth || subgroupConflict))

		if current.LessonNumber == item.LessonNumber && conflict {
			return fmt.Errorf("%w: duplicate lesson for this weekday and subgroup on week %v", ErrItemConflict, current.Weektype)
		}
	}

	return nil
}

type CalendarSchedule struct {
	Items []ScheduleItem
}

// CalendarScheduleFromCycled returns Calendar Schedule based on Cycled schedule Start and End dates
func CalendarScheduleFromCycled(eduGroupID uuid.UUID, semester int, cycled *CycledSchedule, educationStartDate time.Time) (*Schedule, error) {
	id := uuid.New()

	if cycled == nil {
		return nil, errors.New("nil cycled schedule")
	}

	svc := NewScheduleService()

	var items []ScheduleItem
	for d := cycled.StartDate; !d.After(cycled.EndDate); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Sunday {
			continue
		}

		dateItems, err := svc.ListScheduleItemByDate(cycled, educationStartDate, d)
		if err != nil {
			return nil, fmt.Errorf("get item for date %s error: %w", d.Format(time.DateOnly), err)
		}

		if len(dateItems) > 0 {
			items = append(items, dateItems...)
		}
	}

	return &Schedule{
		ID:         id,
		EduGroupID: eduGroupID,
		Semester:   semester,
		Type:       ScheduleTypeCalendar,
		Calendar: &CalendarSchedule{
			Items: items,
		},
	}, nil
}

// ListItem returns ScheduleItem array
func (s CalendarSchedule) ListItem() []ScheduleItem {
	return s.Items
}

// AddItem
func (s *CalendarSchedule) AddItem(
	discipline string,
	teacherID uuid.UUID,
	date time.Time,
	studentsCount int16,
	lessonNumber int8,
	subgroup int8,
	weeknum int,
	lessonType int8,
	classroom string,
) error {
	var argErr error

	if lessonNumber < 0 {
		argErr = errors.Join(argErr, errors.New("invalid lesson number"))
	}

	if subgroup < 0 {
		argErr = errors.Join(argErr, errors.New("invalid subgroup"))
	}

	if date.IsZero() {
		argErr = errors.Join(argErr, errors.New("invalid date value"))
	}

	if weeknum < 1 {
		argErr = errors.Join(argErr, errors.New("invalid weeknum value"))
	}

	y, m, d := date.Date()
	date = time.Date(y, m, d, 0, 0, 0, 0, date.Location())

	weekday := date.Weekday()
	if weekday == time.Sunday {
		argErr = errors.Join(argErr, errors.New("schedule item can not be added for sunday"))
	}

	lt, err := NewItemLessonType(lessonType)
	if err != nil {
		argErr = errors.Join(argErr, err)
	}

	cr, err := NewClassroom(classroom)
	if err != nil {
		argErr = errors.Join(argErr, err)
	}

	if argErr != nil {
		return argErr
	}

	item := ScheduleItem{
		Discipline:    discipline,
		TeacherID:     teacherID,
		Date:          &date,
		StudentsCount: studentsCount,
		Weekday:       weekday,
		LessonNumber:  lessonNumber,
		Subgroup:      subgroup,
		Weeknum:       &weeknum,
		LessonType:    lt,
		Classroom:     cr,
	}

	if vErr := s.validateItem(&item); vErr != nil {
		return vErr
	}

	s.Items = append(s.Items, item)

	return nil
}

// RemoveItem
func (s *CalendarSchedule) RemoveItem(date time.Time, lessonNumber, subgroup int8) error {
	var argErr error

	if lessonNumber < 0 {
		argErr = errors.Join(argErr, errors.New("invalid lesson number"))
	}

	if subgroup < 0 {
		argErr = errors.Join(argErr, errors.New("invalid subgroup"))
	}

	if argErr != nil {
		return argErr
	}

	idx := slices.IndexFunc(s.Items, func(item ScheduleItem) bool {
		return item.Date.Equal(date) && item.LessonNumber == lessonNumber && item.Subgroup == subgroup
	})

	if idx < 0 {
		return errors.New("item not found")
	}

	s.Items = append(s.Items[:idx], s.Items[idx+1:]...)

	return nil
}

func (s *CalendarSchedule) validateItem(item *ScheduleItem) error {
	if item.Weeknum == nil {
		return fmt.Errorf("weeknum can not be empty in calendar schedule")
	}

	if item.Date == nil {
		return fmt.Errorf("date can not be empty in calendar schedule")
	}

	for _, current := range s.Items {
		if current.Date.Equal(*item.Date) && current.LessonNumber == item.LessonNumber && current.Subgroup == item.Subgroup {
			return fmt.Errorf("%w: duplicate lesson for this subgroup on date %s", ErrItemConflict, item.Date.Format(time.DateOnly))
		}
	}

	return nil
}
