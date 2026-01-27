package schedules

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Weektype int8

const (
	WeekTypeUneven Weektype = iota
	WeekTypeEven
	WeekTypeBoth
)

var weektypeNames = []string{
	"odd",
	"even",
	"both",
}

func (w Weektype) String() string {
	return weektypeNames[w]
}

func NewWeekType(wt int8) (Weektype, error) {
	if int(wt) < 0 || int(wt) >= len(weektypeNames) {
		return 0, errors.New("unknown week type")
	}

	return Weektype(wt), nil
}

type ItemLessonType int8

const (
	ItemTypeLecture ItemLessonType = iota
	ItemTypePractice
	ItemTypeSeminar
	ItemTypeExam
	ItemTypeLaboratory
)

var lessonTypeNames = []string{
	"lecture",
	"practice",
	"seminar",
	"exam",
	"laboratory",
}

func (t ItemLessonType) String() string {
	return lessonTypeNames[t]
}

func NewItemLessonType(t int8) (ItemLessonType, error) {
	if int(t) < 0 || int(t) >= len(lessonTypeNames) {
		return 0, errors.New("unknown item type")
	}

	return ItemLessonType(t), nil
}

type Cabinet struct {
	Auditorium string
	Building   string
}

type ScheduleItem struct {
	Discipline    string
	TeacherID     uuid.UUID
	Weekday       time.Weekday
	StudentsCount int16
	Date          *time.Time
	LessonNumber  int8
	Subgroup      int8
	Weektype      *Weektype
	Weeknum       *int
	LessonType    ItemLessonType
	Cabinet       Cabinet
}
