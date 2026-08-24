package usecases

import (
	"context"
	"log/slog"
	"schedule-generator/internal/application/services"
	"schedule-generator/internal/common"
	"schedule-generator/internal/domain/schedules"
	"schedule-generator/internal/infrastructure/db/postgres/repository"
	"schedule-generator/pkg/execerror"
	"time"
)

var weekdaynames = map[time.Weekday]string{
	time.Monday:    "monday",
	time.Tuesday:   "tuesday",
	time.Wednesday: "wednesday",
	time.Thursday:  "thursday",
	time.Friday:    "friday",
	time.Saturday:  "saturday",
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

type CabinetWorkloadBuilding = map[string]CabinetWorkloadDay

// type CabinetWorkloadOutput = map[string]CabinetWorkloadBuilding

type CabinetWorkloadOutput struct {
	AcademicYearStart          string                             `json:"academic_year_start"`
	MaxPairsPerDay             string                             `json:"max_pairs_per_day"`
	CabinetWorkloadFinalOutput map[string]CabinetWorkloadBuilding `json:"cabinet_workload_final_output"`
}

type CabinetWorkloadUsecaseRepo interface {
	ListCycledCabinetWorkload(ctx context.Context) ([]repository.CabinetWorkloadItem, error)
	ListCycledPractices(ctx context.Context) ([]repository.CabinetWorkloadPractice, error)
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

// версия с проверкой пользователя// func (uc *CabinetWorkloadUsecase) GetCabinetWorkload(ctx context.Context, user *users.User) (CabinetWorkloadOutput, error) {
// версия с отдачей просто мапы // func (uc *CabinetWorkloadUsecase) GetCabinetWorkload(ctx context.Context) (CabinetWorkloadOutput, error) {
func (uc *CabinetWorkloadUsecase) GetCabinetWorkload(ctx context.Context) (*CabinetWorkloadOutput, error) {
	// if user == nil {
	// 	return nil, execerror.NewExecError(execerror.TypeForbbiden, nil)
	// }

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

	practiceIndex := make(map[string][]WorkloadPractice, len(practiceItems))
	for _, p := range practiceItems {
		pracType := schedules.PracticeType(p.PracticeType)
		practice := WorkloadPractice{
			PracticeType: pracType.String(),
			StartDate:    p.StartDate.In(common.DefaultTimezone).Format(time.DateOnly),
			EndDate:      p.EndDate.In(common.DefaultTimezone).Format(time.DateOnly),
			Group:        p.EduGroupNumber,
		}

		practiceIndex[p.EduGroupNumber] = append(practiceIndex[p.EduGroupNumber], practice)
	}

	var result CabinetWorkloadOutput = CabinetWorkloadOutput{
		AcademicYearStart:          time.Now().Format("2026-12-25"),
		MaxPairsPerDay:             "7",
		CabinetWorkloadFinalOutput: make(map[string]CabinetWorkloadBuilding),
	}

	var flagOfCopyLesson bool
	for _, item := range items {
		flagOfCopyLesson = false

		weektypeStr := "both"
		if item.Weektype != nil {
			wt := schedules.Weektype(*item.Weektype)
			weektypeStr = wt.String()
		}

		lessType := schedules.ItemLessonType(item.LessonType)
		lessonTypeStr := lessType.String()

		dayName, ok := weekdaynames[item.Weekday]
		if !ok {
			uc.logger.Warn("Unexpected weekday in workload item", "weekday", item.Weekday)
			continue
		}

		groupPractices := practiceIndex[item.EduGroupNumber]
		if groupPractices == nil {
			groupPractices = []WorkloadPractice{}
		}

		var groups []string = []string{item.EduGroupNumber}
		lesson := CabinetWorkloadLesson{
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
			result.CabinetWorkloadFinalOutput[building] = make(CabinetWorkloadBuilding)
		}

		if _, exists := result.CabinetWorkloadFinalOutput[building][auditorium]; !exists {
			result.CabinetWorkloadFinalOutput[building][auditorium] = make(CabinetWorkloadDay)
		}

		// TODO разобраться как помечаются группы/подгруппы, какие "все", какие "п/г-1"
		for i := 0; i < len(result.CabinetWorkloadFinalOutput[building][auditorium][dayName]); i++ {

			if result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].LessonNumber == lesson.LessonNumber &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].WeekType == lesson.WeekType &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].Discipline == lesson.Discipline &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].TeacherName == lesson.TeacherName &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].EduGroup[0] != lesson.EduGroup[0] &&
				result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i].LessonType == lesson.LessonType {

				var v *CabinetWorkloadLesson = &result.CabinetWorkloadFinalOutput[building][auditorium][dayName][i]
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
