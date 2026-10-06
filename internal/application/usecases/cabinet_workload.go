package usecases

import (
	"context"
	"log/slog"
	"schedule-generator/internal/application/services"
	"schedule-generator/internal/common"
	cabinetworkload "schedule-generator/internal/domain/cabinet_workload"
	"schedule-generator/internal/domain/schedules"
	"schedule-generator/internal/domain/users"
	"schedule-generator/pkg/execerror"
	"strings"
	"time"
)

type CabinetWorkloadUsecaseRepo interface {
	ListCycledCabinetWorkload(ctx context.Context) ([]cabinetworkload.CabinetWorkloadItem, error)
	ListCycledPractices(ctx context.Context) ([]cabinetworkload.CabinetWorkloadPractice, error)
}

type CabinetWorkloadUsecase struct {
	authSvc *services.AuthorizationService
	repo    CabinetWorkloadUsecaseRepo
	logger  *slog.Logger
}

func NewCabinetWorkloadUsecase(
	authSvc *services.AuthorizationService,
	repo CabinetWorkloadUsecaseRepo,
	logger *slog.Logger,
) *CabinetWorkloadUsecase {
	var cabWorkloadUsecase *CabinetWorkloadUsecase = &CabinetWorkloadUsecase{
		authSvc: authSvc,
		repo:    repo,
		logger:  logger,
	}

	return cabWorkloadUsecase
}

type CabinetWorkloadOutput struct {
	AcademicYearStart          string
	MaxPairsPerDay             string
	CabinetWorkloadFinalOutput map[string]cabinetworkload.CabinetWorkloadBuilding
}

// GetCabinetWorkload
func (uc *CabinetWorkloadUsecase) GetCabinetWorkload(ctx context.Context, user *users.User) (*CabinetWorkloadOutput, error) {
	if user == nil {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, nil)
	}

	items, err := uc.repo.ListCycledCabinetWorkload(ctx)
	if err != nil {
		uc.logger.Error("ListCycledCabinetWorload error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	practiceItems, err := uc.repo.ListCycledPractices(ctx)
	if err != nil {
		uc.logger.Error("ListCycledPractices error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	practiceIndex := make(map[string][]cabinetworkload.WorkloadPractice, len(practiceItems))
	for _, p := range practiceItems {
		pracType := schedules.PracticeType(p.PracticeType)
		practice := cabinetworkload.WorkloadPractice{
			PracticeType: pracType.String(),
			StartDate:    common.NormalizeTimezone(p.StartDate),
			EndDate:      common.NormalizeTimezone(p.EndDate),
			Group:        p.EduGroupNumber,
		}

		practiceIndex[p.EduGroupNumber] = append(practiceIndex[p.EduGroupNumber], practice)
	}

	now := time.Now()
	year := now.Year()
	if now.Month() < time.September {
		year--
	}
	academicYearStart := time.Date(year, time.September, 0, 0, 0, 0, 0, time.UTC)

	result := CabinetWorkloadOutput{
		AcademicYearStart:          academicYearStart.Format("2006-01-02"),
		MaxPairsPerDay:             "7",
		CabinetWorkloadFinalOutput: make(map[string]cabinetworkload.CabinetWorkloadBuilding),
	}

	var flagOfCopyLesson bool
	for _, item := range items {
		flagOfCopyLesson = false

		weektypeStr := schedules.WeekTypeBoth.String()
		if item.Weektype != nil {
			wt := schedules.Weektype(*item.Weektype)
			weektypeStr = wt.String()
		}

		lessType := schedules.ItemLessonType(item.LessonType)
		lessonTypeStr := lessType.String()

		if item.Weekday == time.Sunday {
			uc.logger.Warn("Unexpected weekday in workload item", "weekday", item.Weekday)
			continue
		}
		dayName := strings.ToLower(item.Weekday.String())

		groupPractices := practiceIndex[item.EduGroupNumber]
		if groupPractices == nil {
			groupPractices = []cabinetworkload.WorkloadPractice{}
		}

		var groups []string = []string{item.EduGroupNumber}
		lesson := cabinetworkload.CabinetWorkloadLesson{
			LessonNumber:  item.LessonNumber,
			WeekType:      weektypeStr,
			Discipline:    item.Discipline,
			TeacherName:   item.TeacherName,
			EduGroup:      groups,
			StudentsCount: item.StudentsCount,
			LessonType:    lessonTypeStr,
			Subgroup:      item.Subgroup,
			Practices:     groupPractices,
		}

		building := item.CabinetBuilding
		auditorium := item.CabinetAuditorium

		if _, exists := result.CabinetWorkloadFinalOutput[building]; !exists {
			result.CabinetWorkloadFinalOutput[building] = make(cabinetworkload.CabinetWorkloadBuilding)
		}

		if _, exists := result.CabinetWorkloadFinalOutput[building][auditorium]; !exists {
			result.CabinetWorkloadFinalOutput[building][auditorium] = make(cabinetworkload.CabinetWorkloadDay)
		}

		for i := 0; i < len(result.CabinetWorkloadFinalOutput[building][auditorium][dayName]); i++ {

			if result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].LessonNumber == lesson.LessonNumber &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].WeekType == lesson.WeekType &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].Discipline == lesson.Discipline &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].TeacherName == lesson.TeacherName &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].EduGroup[0] != lesson.EduGroup[0] &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].LessonType == lesson.LessonType {

				var v *cabinetworkload.CabinetWorkloadLesson = &result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i]
				v.EduGroup = append(v.EduGroup, lesson.EduGroup...)
				v.StudentsCount += lesson.StudentsCount
				v.Subgroup = 0
				v.Practices = append(v.Practices, lesson.Practices...)
				flagOfCopyLesson = true
			}
		}

		if !flagOfCopyLesson {
			result.CabinetWorkloadFinalOutput[building][auditorium][dayName] = append(result.CabinetWorkloadFinalOutput[building][auditorium][dayName], lesson)
		}

	}

	return &result, nil
}
