package handler

import (
	"context"
	"net/http"
	"schedule-generator/internal/application/usecases"
	cabinetworkload "schedule-generator/internal/domain/cabinet_workload"
	"schedule-generator/internal/domain/users"

	"github.com/labstack/echo/v4"
)

type CabinetWorkloadUsecase interface {
	GetCabinetWorkload(ctx context.Context, user *users.User) (*usecases.CabinetWorkloadOutput, error)
}

type WorkloadPractice struct {
	PracticeType string `json:"practice_type"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Group        string `json:"group"`
}

type CabinetWorkloadLesson struct {
	LessonNumber  int8               `json:"lesson_number"`
	WeekType      string             `json:"weektype"`
	Discipline    string             `json:"discipline"`
	TeacherName   string             `json:"teacher_name"`
	EduGroup      []string           `json:"edu_group"`
	StudentsCount int16              `json:"students_count"`
	LessonType    string             `json:"lesson_type"`
	Subgroup      int8               `json:"subgroup"`
	Practices     []WorkloadPractice `json:"practices"`
}

type CabinetWorkloadDay = map[string][]CabinetWorkloadLesson

type CabinetWorkloadBuilding map[string]CabinetWorkloadDay

type CabinetWorkloadOutput struct {
	AcademicYearStart          string                             `json:"academic_year_start"`
	MaxPairsPerDay             string                             `json:"max_pairs_per_day"`
	CabinetWorkloadFinalOutput map[string]CabinetWorkloadBuilding `json:"cabinet_workload_final_output"`
}

func (h *Handler) GetCabinetWorkload(c echo.Context) error {
	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	out, err := h.cabinetWorkload.GetCabinetWorkload(c.Request().Context(), user)
	if err != nil {
		h.logger.Error("GetCabinetWorkload error", "error", err)
		return err
	}

	// lengthOfWKLFinalOutput := len(out.CabinetWorkloadFinalOutput)
	result := CabinetWorkloadOutput{
		AcademicYearStart: out.AcademicYearStart,
		MaxPairsPerDay:    out.MaxPairsPerDay,
		// CabinetWorkloadFinalOutput: make(map[string]CabinetWorkloadBuilding, lengthOfWKLFinalOutput),
	}

	result.CabinetWorkloadFinalOutput = cabinetWorkloadOutputToView(out.CabinetWorkloadFinalOutput)

	return WrapResponse(http.StatusOK, result).Send(c)
}

func cabinetWorkloadOutputToView(data map[string]cabinetworkload.CabinetWorkloadBuilding) map[string]CabinetWorkloadBuilding {
	result := make(map[string]CabinetWorkloadBuilding, len(data))

	for keyOfBuilding, buildingMap := range data {

		if _, exists := result[keyOfBuilding]; !exists {
			result[keyOfBuilding] = make(CabinetWorkloadBuilding, len(buildingMap))
		}

		for keyOfCabinet, dayMap := range buildingMap {

			if _, exists := result[keyOfBuilding][keyOfCabinet]; !exists {
				result[keyOfBuilding][keyOfCabinet] = make(CabinetWorkloadDay, len(dayMap))
			}

			for keyOfDay, lessonSl := range dayMap {

				if _, exists := result[keyOfBuilding][keyOfCabinet][keyOfDay]; !exists {
					// result[keyOfBuilding][keyOfCabinet][keyOfDay] = make([]CabinetWorkloadLesson, 0, len(lessonSl))
					result[keyOfBuilding][keyOfCabinet][keyOfDay] = make([]CabinetWorkloadLesson, len(lessonSl))

				}

				for indexOfLesson, lesson := range lessonSl {

					practicesHand := make([]WorkloadPractice, len(lesson.Practices))
					for i, practice := range lesson.Practices {
						practicesHand[i] = WorkloadPractice{
							PracticeType: practice.PracticeType,
							StartDate:    practice.StartDate,
							EndDate:      practice.EndDate,
							Group:        practice.Group,
						}
					}

					newLesson := CabinetWorkloadLesson{
						LessonNumber:  lesson.LessonNumber,
						WeekType:      lesson.WeekType,
						Discipline:    lesson.Discipline,
						TeacherName:   lesson.TeacherName,
						EduGroup:      lesson.EduGroup,
						StudentsCount: lesson.StudentsCount,
						LessonType:    lesson.LessonType,
						Subgroup:      lesson.Subgroup,
						Practices:     practicesHand,
					}

					result[keyOfBuilding][keyOfCabinet][keyOfDay][indexOfLesson] = newLesson
				}
			}
		}
	}

	return result
}
