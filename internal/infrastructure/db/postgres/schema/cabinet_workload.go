package schema

import (
	cabinetworkload "schedule-generator/internal/domain/cabinet_workload"
	"time"
)

type CabinetWorkloadItem struct {
	Discipline        string       `gorm:"column:discipline"`
	Weekday           time.Weekday `gorm:"column:weekday"`
	LessonNumber      int8         `gorm:"column:lesson_number"`
	Subgroup          int8         `gorm:"column:subgroup"`
	Weektype          *int8        `gorm:"column:weektype"`
	LessonType        int8         `gorm:"column:lesson_type"`
	StudentsCount     int16        `gorm:"column:students_count"`
	CabinetAuditorium string       `gorm:"column:cabinet_auditorium"`
	CabinetBuilding   string       `gorm:"column:cabinet_building"`

	TeacherName string `gorm:"column:teacher_name"`

	EduGroupNumber string `gorm:"column:edu_group_number"`
}

func CabinetWorkloadItemToSchema(c *cabinetworkload.CabinetWorkloadItem) *CabinetWorkloadItem {
	cabWLITS := CabinetWorkloadItem{
		Discipline:        c.Discipline,
		Weekday:           c.Weekday,
		LessonNumber:      c.LessonNumber,
		Subgroup:          c.Subgroup,
		Weektype:          c.Weektype,
		LessonType:        c.LessonType,
		StudentsCount:     c.StudentsCount,
		CabinetAuditorium: c.CabinetAuditorium,
		CabinetBuilding:   c.CabinetBuilding,
		TeacherName:       c.TeacherName,
		EduGroupNumber:    c.EduGroupNumber,
	}

	return &cabWLITS
}

func CabinetWorkloadItemFromSchema(schema *CabinetWorkloadItem) *cabinetworkload.CabinetWorkloadItem {
	cabWLIFS := cabinetworkload.CabinetWorkloadItem{
		Discipline:        schema.Discipline,
		Weekday:           schema.Weekday,
		LessonNumber:      schema.LessonNumber,
		Subgroup:          schema.Subgroup,
		Weektype:          schema.Weektype,
		LessonType:        schema.LessonType,
		StudentsCount:     schema.StudentsCount,
		CabinetAuditorium: schema.CabinetAuditorium,
		CabinetBuilding:   schema.CabinetBuilding,
		TeacherName:       schema.CabinetBuilding,
		EduGroupNumber:    schema.EduGroupNumber,
	}

	return &cabWLIFS
}

type CabinetWorkloadPractice struct {
	PracticeType   int8      `gorm:"column:practice_type"`
	StartDate      time.Time `gorm:"column:start_date"`
	EndDate        time.Time `gorm:"column:end_date"`
	EduGroupNumber string    `gorm:"column:edu_group_number"`
}

func CabinetWorkloadPracticeToSchema(p *cabinetworkload.CabinetWorkloadPractice) *CabinetWorkloadPractice {
	cabWLPTS := CabinetWorkloadPractice{
		PracticeType:   p.PracticeType,
		StartDate:      p.StartDate,
		EndDate:        p.EndDate,
		EduGroupNumber: p.EduGroupNumber,
	}

	return &cabWLPTS
}

func CabinetWorkloadPracticeFromSchema(schema *CabinetWorkloadPractice) *cabinetworkload.CabinetWorkloadPractice {
	cabWLPFS := cabinetworkload.CabinetWorkloadPractice{
		PracticeType:   schema.PracticeType,
		StartDate:      schema.StartDate,
		EndDate:        schema.EndDate,
		EduGroupNumber: schema.EduGroupNumber,
	}

	return &cabWLPFS
}
