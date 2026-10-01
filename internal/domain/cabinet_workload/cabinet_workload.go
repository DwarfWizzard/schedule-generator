package cabinetworkload

import (
	"time"
)

type CabinetWorkloadItem struct {
	Discipline        string
	Weekday           time.Weekday
	LessonNumber      int8
	Subgroup          int8
	Weektype          *int8
	LessonType        int8
	StudentsCount     int16
	CabinetAuditorium string
	CabinetBuilding   string

	TeacherName string

	EduGroupNumber string
}

type CabinetWorkloadPractice struct {
	PracticeType   int8
	StartDate      time.Time
	EndDate        time.Time
	EduGroupNumber string
}

type WorkloadPractice struct {
	PracticeType string
	StartDate    string
	EndDate      string
	Group        string
}

type CabinetWorkloadLesson struct {
	LessonNumber  int8
	WeekType      string
	Discipline    string
	TeacherName   string
	EduGroup      []string
	StudentsCount int16
	LessonType    string
	Subgroup      int8
	Practices     []WorkloadPractice
}

type CabinetWorkloadDay map[string][]CabinetWorkloadLesson

type CabinetWorkloadBuilding map[string]CabinetWorkloadDay
