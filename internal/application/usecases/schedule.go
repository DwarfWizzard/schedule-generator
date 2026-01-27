package usecases

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strconv"
	"time"

	"schedule-generator/internal/application/acl/exporter"
	"schedule-generator/internal/domain/cabinets"
	edugroups "schedule-generator/internal/domain/edu_groups"
	"schedule-generator/internal/domain/schedules"
	"schedule-generator/internal/domain/teachers"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type ScheduleUsecaseRepo interface {
	schedules.Repository
	edugroups.Repository
	cabinets.Repository

	GetScheduleByEduGroupIDAndSemester(ctx context.Context, eduGroupID uuid.UUID, semester int) (*schedules.Schedule, error)
	MapEduGroupsBySchedules(ctx context.Context, scheduleIDs uuid.UUIDs) (map[uuid.UUID]edugroups.EduGroup, error)
	MapTeacherByIDs(ctx context.Context, teacherIDs uuid.UUIDs) (map[uuid.UUID]teachers.Teacher, error)

	db.TransactionalRepository
}

type ScheduleUsecase struct {
	repo     ScheduleUsecaseRepo
	exporter exporter.Factory
	logger   *slog.Logger
}

func NewScheduleUsecase(repo ScheduleUsecaseRepo, exporter exporter.Factory, logger *slog.Logger) *ScheduleUsecase {
	return &ScheduleUsecase{
		repo:     repo,
		exporter: exporter,
		logger:   logger,
	}
}

type ScheduleItemDTO struct {
	schedules.ScheduleItem
	TeacherName string
}

type ScheduleDTO struct {
	ID         uuid.UUID
	EduGroupID uuid.UUID
	Semester   int
	Type       schedules.ScheduleType
	StartDate  *time.Time
	EndDate    *time.Time
	Items      []ScheduleItemDTO
}

type CreateScheduleInput struct {
	EduGroupID uuid.UUID
	Semester   int
	StartDate  *time.Time
	EndDate    *time.Time
}

type CreateScheduleOutput struct {
	ScheduleDTO
	EduGroupNumber string
}

// CreateSchedule
func (uc *ScheduleUsecase) CreateSchedule(ctx context.Context, input CreateScheduleInput) (*CreateScheduleOutput, error) {
	logger := uc.logger.With("edu_group_id", input.EduGroupID)

	group, err := uc.repo.GetEduGroup(ctx, input.EduGroupID)
	if err != nil {
		logger.Error("Get edu group error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu group not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if _, err := uc.repo.GetScheduleByEduGroupIDAndSemester(ctx, input.EduGroupID, input.Semester); err == nil {
		return nil, execerror.NewExecError(execerror.TypeProcessingConflict, errors.New("schedule for group and semester already exists")).
			AddDetails("edu_group_id", input.EduGroupID.String()).
			AddDetails("semester", strconv.FormatInt(int64(input.Semester), 10))
	} else if !errors.Is(err, db.ErrorNotFound) {
		logger.Error("Check if schedule alreay exists for semeter error", "error", err, "semester", input.Semester)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if input.StartDate == nil || input.EndDate == nil {
		return nil, execerror.NewExecError(execerror.TypeUnimpemented, errors.New("calendar schedule not implemented"))
	}

	schedule, err := schedules.NewCycledSchedule(input.EduGroupID, input.Semester, *input.StartDate, *input.EndDate)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveSchedule(ctx, schedule)
	if err != nil {
		logger.Error("Save schedule error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	dto, _ := scheduleToCycledScheduleDTO(schedule, nil, false)

	return &CreateScheduleOutput{
		ScheduleDTO:    dto,
		EduGroupNumber: group.Number,
	}, nil
}

type GetScheduleOutput struct {
	ScheduleDTO
	EduGroupNumber string
}

// GetSchedule
func (uc *ScheduleUsecase) GetSchedule(ctx context.Context, scheduleID uuid.UUID) (*GetScheduleOutput, error) {
	logger := uc.logger

	schedule, err := uc.repo.GetSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Get schedule error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("schedule not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	group, err := uc.repo.GetEduGroup(ctx, schedule.EduGroupID)
	if err != nil {
		logger.Error("Get schedules group error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	var teacherIDs uuid.UUIDs
	m := make(map[uuid.UUID]struct{})

	for _, item := range schedule.ListItem() {
		if _, ok := m[item.TeacherID]; ok {
			continue
		}

		teacherIDs = append(teacherIDs, item.TeacherID)
	}

	teachersMap, err := uc.repo.MapTeacherByIDs(ctx, teacherIDs)
	if err != nil {
		logger.Error("Get teachers map error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	dto, err := scheduleToCycledScheduleDTO(schedule, teachersMap, true)
	if err != nil {
		logger.Error("Create schedule dto error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetScheduleOutput{ScheduleDTO: dto, EduGroupNumber: group.Number}, nil
}

type ListScheduleOutput = []GetScheduleOutput

// ListSchedule
func (uc *ScheduleUsecase) ListSchedule(ctx context.Context) (ListScheduleOutput, error) {
	logger := uc.logger

	schedules, err := uc.repo.ListSchedule(ctx)
	if err != nil {
		logger.Error("Get list schedule error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	scheduleIDs := make(uuid.UUIDs, len(schedules))
	for i, schedule := range schedules {
		scheduleIDs[i] = schedule.ID
	}

	groups, err := uc.repo.MapEduGroupsBySchedules(ctx, scheduleIDs)
	if err != nil {
		logger.Error("Map edu groups by schedules error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make(ListScheduleOutput, len(schedules))
	for idx, schedule := range schedules {
		group, ok := groups[schedule.EduGroupID]
		if !ok {
			logger.Error(fmt.Sprintf("Edu group for schedule %s not found", group.ID))
			return nil, execerror.NewExecError(execerror.TypeInternal, nil)
		}

		dto, _ := scheduleToCycledScheduleDTO(&schedule, nil, false)

		result[idx] = GetScheduleOutput{ScheduleDTO: dto, EduGroupNumber: group.Number}
	}

	return result, nil
}

type AddItemToScheduleInput struct {
	Discipline    string
	TeacherID     uuid.UUID
	CabinetID     uuid.UUID
	StudentsCount int16
	Date          *time.Time
	Weeknum       *int
	Weekday       *time.Weekday
	Weektype      *int8
	LessonNumber  int8
	Subgroup      int8
	LessonType    int8
}

// AddItemToSchedule
func (uc *ScheduleUsecase) AddItemsToSchedule(ctx context.Context, scheduleID uuid.UUID, input []AddItemToScheduleInput) error {
	logger := uc.logger.With("schedule_id", scheduleID)

	tx, rollback, commit, err := uc.repo.AsTransaction(ctx, db.IsoLevelDefault)
	if err != nil {
		logger.Error("Start transaction error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}
	defer rollback(ctx)

	repo := tx.(ScheduleUsecaseRepo)

	schedule, err := repo.GetSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Get schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	//TODO: handle calendar schedule
	if schedule.Type != schedules.ScheduleTypeCycled {
		logger.Error("Schedule is not cycled")
		return execerror.NewExecError(execerror.TypeUnimpemented, errors.New("currently supproted schedule is cycled"))
	}

	for i, item := range input {
		if item.Weekday == nil {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing weekday"))
		}

		if item.Weektype == nil {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing weektype"))
		}

		cabinet, err := repo.GetCabinet(ctx, item.CabinetID)
		if err != nil {
			logger.Error("Get cabinet error", "error", err)
			if errors.Is(err, db.ErrorNotFound) {
				return execerror.NewExecError(execerror.TypeInvalidInput, fmt.Errorf("cabinet %s not found", item.CabinetID))
			}

			return execerror.NewExecError(execerror.TypeInternal, nil)
		}

		cabinetValue := schedules.Cabinet{
			Building:   cabinet.Building,
			Auditorium: cabinet.Auditorium,
		}

		err = schedule.Cycled.AddItem(
			item.Discipline,
			item.TeacherID,
			*item.Weekday,
			item.StudentsCount,
			item.LessonNumber,
			item.Subgroup,
			*item.Weektype,
			item.LessonType,
			cabinetValue,
		)
		if err != nil {
			return execerror.NewExecError(execerror.TypeInvalidInput, err).AddDetails("input_idx", strconv.FormatInt(int64(i), 10))
		}
	}

	err = uc.repo.SaveSchedule(ctx, schedule)
	if err != nil {
		logger.Error("Save schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	err = commit(ctx)
	if err != nil {
		logger.Error("Save updated schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}

// GetListScheduleItemForSpecifiedDate
func (uc *ScheduleUsecase) GetListScheduleItemForSpecifiedDate(ctx context.Context, scheduleID uuid.UUID, date time.Time) ([]schedules.ScheduleItem, error) {
	logger := uc.logger.With("schedule_id", scheduleID, "date", date)

	schedule, err := uc.repo.GetSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Get schedule error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("schedule not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if schedule.Type != schedules.ScheduleTypeCycled {
		logger.Error("Schedule is not cycled")
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("allowed only for cycled schedule"))
	}

	group, err := uc.repo.GetEduGroup(ctx, schedule.EduGroupID)
	if err != nil {
		logger.Error("Get schedule edu group error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	educationStartDate := group.GetEducationStartDateBySemester(schedule.Semester)

	scheduleSvc := schedules.NewScheduleService()

	items, err := scheduleSvc.ListScheduleItemByDate(schedule.Cycled, educationStartDate, date)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	return items, nil
}

// ExportSchedule
func (uc *ScheduleUsecase) ExportSchedule(ctx context.Context, scheduleID uuid.UUID, format string, dst io.Writer) error {
	logger := uc.logger.With("schedule_id", scheduleID)

	schedule, err := uc.repo.GetSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Get schedule error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("schedule not found"))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	exp, err := uc.exporter.ByFormat(format)
	if err != nil {
		logger.Error("Get exporter by formate error", "error", err)
		if errors.Is(err, exporter.ErrUnknownFormat) {
			return execerror.NewExecError(execerror.TypeInvalidInput, err)
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	err = exp.Export(ctx, schedule, dst)
	if err != nil {
		logger.Error("Export schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}

// ExportCycledScheduleAsCalendar
func (uc *ScheduleUsecase) ExportCycledScheduleAsCalendar(ctx context.Context, scheduleID uuid.UUID, format string, dst io.Writer) error {
	logger := uc.logger.With("schedule_id", scheduleID)

	schedule, err := uc.repo.GetSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Get schedule error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("schedule not found"))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if schedule.Type != schedules.ScheduleTypeCycled {
		return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("schedule is not cycled"))
	}

	group, err := uc.repo.GetEduGroup(ctx, schedule.EduGroupID)
	if err != nil {
		logger.Error("Get schedule edu group error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	calendarSchedule, err := schedules.CalendarScheduleFromCycled(schedule.EduGroupID, schedule.Semester, schedule.Cycled, group.GetEducationStartDateBySemester(schedule.Semester))
	if err != nil {
		logger.Error("Make calendar from cycled schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	log.Println(calendarSchedule.ListItem())

	exp, err := uc.exporter.ByFormat(format)
	if err != nil {
		logger.Error("Get exporter by formate error", "error", err)
		if errors.Is(err, exporter.ErrUnknownFormat) {
			return execerror.NewExecError(execerror.TypeInvalidInput, err)
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	err = exp.Export(ctx, calendarSchedule, dst)
	if err != nil {
		logger.Error("Export schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}

type RemoveItemFromScheduleInput struct {
	Date         *time.Time
	Weekday      *time.Weekday
	LessonNumber int8
	Subgroup     int8
	Weektype     *int8
}

// RemoveItemsFromSchedule
func (uc *ScheduleUsecase) RemoveItemsFromSchedule(ctx context.Context, scheduleID uuid.UUID, input []RemoveItemFromScheduleInput) error {
	logger := uc.logger.With("schedule_id", scheduleID)

	tx, rollback, commit, err := uc.repo.AsTransaction(ctx, db.IsoLevelDefault)
	if err != nil {
		logger.Error("Start transaction error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}
	defer rollback(ctx)

	repo := tx.(ScheduleUsecaseRepo)

	schedule, err := repo.GetSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Get schedule error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("schedule not found"))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	for _, item := range input {
		switch schedule.Type {
		case schedules.ScheduleTypeCycled:
			if item.Weekday == nil {
				return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing weekday"))
			}

			if item.Weektype == nil {
				return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing weektype"))
			}

			err := schedule.Cycled.RemoveItem(*item.Weekday, item.LessonNumber, item.Subgroup, *item.Weektype)
			if err != nil {
				logger.Error("Remove item error", "error", err)
				return execerror.NewExecError(execerror.TypeInvalidInput, err)
			}
		case schedules.ScheduleTypeCalendar:
			if item.Date == nil {
				return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing date"))
			}

			err := schedule.Calendar.RemoveItem(*item.Date, item.LessonNumber, item.Subgroup)
			if err != nil {
				logger.Error("Remove item error", "error", err)
				return execerror.NewExecError(execerror.TypeInvalidInput, err)
			}
		}
	}

	err = repo.SaveSchedule(ctx, schedule)
	if err != nil {
		logger.Error("Save schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	err = commit(ctx)
	if err != nil {
		logger.Error("Save updated schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}

// UpdateItemInSchedule
func (uc *ScheduleUsecase) UpdateItemInSchedule(ctx context.Context, scheduleID uuid.UUID, input AddItemToScheduleInput) error {
	logger := uc.logger.With("schedule_id", scheduleID)

	tx, rollback, commit, err := uc.repo.AsTransaction(ctx, db.IsoLevelDefault)
	if err != nil {
		logger.Error("Start transaction error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}
	defer rollback(ctx)

	repo := tx.(ScheduleUsecaseRepo)

	schedule, err := repo.GetSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Get schedule error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("schedule not found"))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	cabinet, err := repo.GetCabinet(ctx, input.CabinetID)
	if err != nil {
		logger.Error("Get cabinet error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, fmt.Errorf("cabinet %s not found", input.CabinetID))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	cabinetValue := schedules.Cabinet{
		Building:   cabinet.Building,
		Auditorium: cabinet.Auditorium,
	}

	switch schedule.Type {
	case schedules.ScheduleTypeCycled:
		if input.Weekday == nil {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing weekday"))
		}

		if input.Weektype == nil {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing weektype"))
		}

		err := schedule.Cycled.RemoveItem(*input.Weekday, input.LessonNumber, input.Subgroup, *input.Weektype)
		if err != nil {
			logger.Error("Remove item error", "error", err)
			return execerror.NewExecError(execerror.TypeInvalidInput, err)
		}

		err = schedule.Cycled.AddItem(
			input.Discipline,
			input.TeacherID,
			*input.Weekday,
			input.StudentsCount,
			input.LessonNumber,
			input.Subgroup,
			*input.Weektype,
			input.LessonType,
			cabinetValue,
		)
	case schedules.ScheduleTypeCalendar:
		if input.Date == nil {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing date"))
		}

		if input.Weeknum == nil {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("missing weeknum"))
		}

		err := schedule.Calendar.RemoveItem(*input.Date, input.LessonNumber, input.Subgroup)
		if err != nil {
			logger.Error("Remove item error", "error", err)
			return execerror.NewExecError(execerror.TypeInvalidInput, err)
		}

		err = schedule.Calendar.AddItem(
			input.Discipline,
			input.TeacherID,
			*input.Date,
			input.StudentsCount,
			input.LessonNumber,
			input.Subgroup,
			*input.Weeknum,
			input.LessonType,
			cabinetValue,
		)
	}

	err = repo.SaveSchedule(ctx, schedule)
	if err != nil {
		logger.Error("Save schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	err = commit(ctx)
	if err != nil {
		logger.Error("Save updated schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}

// DeleteSchedule
func (uc *ScheduleUsecase) DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	logger := uc.logger

	err := uc.repo.DeleteSchedule(ctx, scheduleID)
	if err != nil {
		logger.Error("Delete schedule error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}

func scheduleToCycledScheduleDTO(schedule *schedules.Schedule, teachersMap map[uuid.UUID]teachers.Teacher, withItems bool) (ScheduleDTO, error) {
	var items []ScheduleItemDTO

	if withItems {
		for _, item := range schedule.Cycled.ListItem() {
			t, ok := teachersMap[item.TeacherID]
			if !ok {
				return ScheduleDTO{}, fmt.Errorf("teacher with id %s for item %s not found", item.TeacherID, item.Discipline)
			}

			items = append(items, ScheduleItemDTO{
				ScheduleItem: item,
				TeacherName:  t.Name,
			})
		}
	}

	dto := ScheduleDTO{
		ID:         schedule.ID,
		Semester:   schedule.Semester,
		EduGroupID: schedule.EduGroupID,
		Items:      items,
	}

	if schedule.Type == schedules.ScheduleTypeCycled {
		dto.StartDate = &schedule.Cycled.StartDate
		dto.EndDate = &schedule.Cycled.EndDate
	}

	return dto, nil
}
