package schedules

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCycledSchedule_NewCycledSchedule(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		schedule, err := NewCycledSchedule(uuid.New(), 0, time.Now(), time.Now().AddDate(0, 0, 1))
		if err != nil {
			t.Fatal(err)
		}

		if schedule.Type != ScheduleTypeCycled {
			t.Errorf("expected schedule type is %v, got: %v", ScheduleTypeCycled, schedule.Type)
		}
	})
	t.Run("fails", func(t *testing.T) {
		groupID := uuid.New()

		cases := map[string]struct {
			semester           int
			startDate, endDate time.Time
		}{
			"invalid semster": {
				semester:  -1,
				startDate: time.Now(),
				endDate:   time.Now().AddDate(0, 0, 1),
			},
			"zero start date": {
				semester:  0,
				startDate: time.Time{},
				endDate:   time.Time{},
			},
			"end date before start": {
				semester:  0,
				startDate: time.Now(),
				endDate:   time.Now().AddDate(0, 0, -1),
			},
		}

		for n, c := range cases {
			t.Run(n, func(t *testing.T) {
				_, err := NewCycledSchedule(groupID, c.semester, c.startDate, c.endDate)
				if err == nil {
					t.Error("unexpected nil value")
				}
			})
		}
	})
}

func TestCycledSchedule_AddItem(t *testing.T) {
	teacherID := uuid.New()
	weekType := WeekTypeBoth

	suitcases := map[string]struct {
		discipline    string
		teacherID     uuid.UUID
		weekday       time.Weekday
		studentsCount int16
		lessonNumber  int8
		subgroup      int8
		weektype      int8
		lessonType    int8
		classroom     string
		result        *ScheduleItem
		err           error
	}{
		"happy-path": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: 0,
			lessonNumber:  0,
			subgroup:      0,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "test",
			result: &ScheduleItem{
				Discipline:    "test",
				TeacherID:     teacherID,
				Weekday:       time.Monday,
				StudentsCount: 0,
				LessonNumber:  0,
				Subgroup:      0,
				Weektype:      &weekType,
				LessonType:    ItemTypeLecture,
				Classroom:     "test",
			},
		},
		"item for sunday": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Sunday,
			studentsCount: 0,
			lessonNumber:  0,
			subgroup:      0,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "test",
			err:           ErrInvalidData,
		},
		"negative students count": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: -1,
			lessonNumber:  0,
			subgroup:      0,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "test",
			err:           ErrInvalidData,
		},
		"negative lesson number": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: 0,
			lessonNumber:  -1,
			subgroup:      0,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "test",
			err:           ErrInvalidData,
		},
		"negative subgroup": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: 0,
			lessonNumber:  0,
			subgroup:      -1,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "test",
			err:           ErrInvalidData,
		},
		"unknown weektype": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: 0,
			lessonNumber:  0,
			subgroup:      0,
			weektype:      int8(len(weektypeNames) + 1),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "test",
			err:           ErrInvalidData,
		},
		"unknown lesson type": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: 0,
			lessonNumber:  0,
			subgroup:      0,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(len(lessonTypeNames) + 1),
			classroom:     "test",
			err:           ErrInvalidData,
		},
		"empty discipline": {
			discipline:    "",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: 0,
			lessonNumber:  0,
			subgroup:      0,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "test",
			err:           ErrInvalidData,
		},
		"empty classroom": {
			discipline:    "test",
			teacherID:     teacherID,
			weekday:       time.Monday,
			studentsCount: 0,
			lessonNumber:  0,
			subgroup:      0,
			weektype:      int8(WeekTypeBoth),
			lessonType:    int8(ItemTypeLecture),
			classroom:     "",
			err:           ErrInvalidData,
		},
	}

	for name, suitcase := range suitcases {
		t.Run(name, func(t *testing.T) {
			schedule, err := NewCycledSchedule(uuid.New(), 0, time.Now(), time.Now())
			if err != nil {
				t.Fatal(err)
			}

			err = schedule.Cycled.AddItem(
				suitcase.discipline,
				suitcase.teacherID,
				suitcase.weekday,
				suitcase.studentsCount,
				suitcase.lessonNumber,
				suitcase.subgroup,
				suitcase.weektype,
				suitcase.lessonType,
				suitcase.classroom,
			)

			if suitcase.err != nil {
				if !errors.Is(err, suitcase.err) {
					t.Fatalf("expected error %v, got %v", suitcase.err, err)
				}

				if len(schedule.Cycled.Items) != 0 {
					t.Fatalf("expected 0 items, got %d", len(schedule.Cycled.Items))
				}

				return
			}

			if len(schedule.Cycled.Items) > 1 {
				t.Fatalf("expected 1 item, got: %d", len(schedule.Cycled.Items))
			}

			items, ok := schedule.Cycled.Items[suitcase.result.Weekday]
			if !ok {
				t.Fatalf("unexpected empty schedule items on %s", suitcase.result.Weekday.String())
			}

			if len(items) == 0 {
				t.Fatalf("unexpected empty schedule items on %s", suitcase.result.Weekday.String())
			}

			if len(items) > 1 {
				t.Fatalf("expected 1 item on %s, got: %d", suitcase.result.Weekday.String(), len(items))
			}

			if !cmpItems(suitcase.result, &items[0]) {
				t.Errorf("expected item is %+v, got: %+v", *suitcase.result, items[0])
			}
		})
	}
}

// func TestCycledSchedule_RemoveItem(t *testing.T) {
// 	suitcases := map[string]struct {
// 		discipline    string
// 		teacherID     uuid.UUID
// 		weekday       time.Weekday
// 		studentsCount int16
// 		lessonNumber  int8
// 		subgroup      int8
// 		weektype      int8
// 		lessonType    int8
// 		classroom     string
// 		result        *ScheduleItem
// 		err           error
// 	}{}
// }

func cmpItems(i1, i2 *ScheduleItem) bool {
	if i1 == nil && i2 == nil {
		return true
	}

	return i1.Discipline == i2.Discipline &&
		i1.TeacherID == i2.TeacherID &&
		i1.Weekday == i2.Weekday &&
		i1.StudentsCount == i2.StudentsCount &&
		i1.LessonNumber == i2.LessonNumber &&
		i1.Subgroup == i2.Subgroup &&
		i1.LessonType == i2.LessonType &&
		i1.Classroom == i2.Classroom &&
		(i1.Date == nil && i2.Date == nil || (i1.Date.Equal(*i2.Date))) &&
		(i1.Weektype == nil && i2.Weektype == nil || (*i1.Weektype == *i2.Weektype)) &&
		(i1.Weeknum == nil && i2.Weeknum == nil || (*i1.Weeknum == *i2.Weeknum))
}
