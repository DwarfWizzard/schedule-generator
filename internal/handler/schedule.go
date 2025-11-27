package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/common"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ScheduleUsecase interface {
	CreateSchedule(ctx context.Context, input usecases.CreateScheduleInput) (*usecases.CreateScheduleOutput, error)
	ListSchedule(ctx context.Context) (usecases.ListScheduleOutput, error)
	GetSchedule(ctx context.Context, scheduleID uuid.UUID) (*usecases.GetScheduleOutput, error)
	AddItemsToSchedule(ctx context.Context, scheduleID uuid.UUID, input []usecases.AddItemToScheduleInput) error
	RemoveItemsFromSchedule(ctx context.Context, scheduleID uuid.UUID, input []usecases.RemoveItemFromScheduleInput) error
	ExportSchedule(ctx context.Context, scheduleID uuid.UUID, format string, dst io.Writer) error
	ExportCycledScheduleAsCalendar(ctx context.Context, scheduleID uuid.UUID, format string, dst io.Writer) error
}

type ScheduleItem struct {
	Discipline    string
	TeacherID     uuid.UUID
	Weekday       string
	StudentsCount int16
	Date          *time.Time
	LessonNumber  int8
	Subgroup      int8
	Weektype      *string
	Weeknum       *int
	LessonType    string
	Classroom     string
}

type Schedule struct {
	ID         uuid.UUID      `json:"id"`
	EduGroupID uuid.UUID      `json:"edu_group_id"`
	Type       string         `json:"type"`
	StartDate  *time.Time     `json:"start_week"`
	EndDate    *time.Time     `json:"end_week"`
	Items      []ScheduleItem `json:"schedule_item"`
}

type CreateScheduleRequest struct {
	//TODO: add calendar
	EduGroupID uuid.UUID `json:"edu_group_id"`
	Semester   int       `json:"semester"`
	StartDate  string    `json:"start_week"`
	EndDate    string    `json:"end_week"`
}

type CreateScheduleResponse struct {
	Schedule
}

// CreateScheduleRequest - POST /v1/schedules
func (h *Handler) CreateSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	var rq CreateScheduleRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	var startDate, endDate *time.Time
	if v, err := time.ParseInLocation(time.DateOnly, rq.StartDate, common.DefaultTimezone); err != nil {
		return ErrInvalidInput
	} else {
		startDate = &v
	}

	if v, err := time.ParseInLocation(time.DateOnly, rq.EndDate, common.DefaultTimezone); err != nil {
		return ErrInvalidInput
	} else {
		endDate = &v
	}

	schedule, err := h.schedule.CreateSchedule(ctx, usecases.CreateScheduleInput{
		EduGroupID: rq.EduGroupID,
		Semester:   rq.Semester,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		h.logger.Error("Create schedule error", "error", err)
		return err
	}

	respData := CreateScheduleResponse{
		Schedule: Schedule{
			ID:         schedule.ID,
			EduGroupID: schedule.EduGroupID,
			StartDate:  schedule.StartDate,
			EndDate:    schedule.EndDate,
			Items:      nil,
		},
	}

	return WrapResponse(http.StatusOK, respData).Send(c)
}

// ListSchedule - GET /v1/schedules
func (h *Handler) ListSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	list, err := h.schedule.ListSchedule(ctx)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	result := make([]Schedule, len(list))

	for idx, schedule := range list {
		result[idx] = scheduleDTOtoView(schedule)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

// GetSchedule - GET /v1/schedules/:id
func (h *Handler) GetSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.schedule.GetSchedule(ctx, scheduleID)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, scheduleDTOtoView(out.ScheduleDTO)).Send(c)
}

type AddScheduleItemRequest struct {
	Discipline    string       `json:"discipline"`
	TeacherID     uuid.UUID    `json:"teacher_id"`
	Weekday       time.Weekday `json:"weekday"`
	StudentsCount int16        `json:"students_count"`
	LessonNumber  int8         `json:"lesson_number"`
	Subgroup      int8         `json:"subgroup"`
	Weektype      int8         `json:"weektype"`
	LessonType    int8         `json:"lesson_type"`
	Classroom     string       `json:"classroom"`
}

// AddScheduleItem - POST /v1/schedules/:id/items
func (h *Handler) AddScheduleItem(c echo.Context) error {
	ctx := c.Request().Context()

	var rq []AddScheduleItemRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	input := make([]usecases.AddItemToScheduleInput, len(rq))

	for i, item := range rq {
		input[i] = usecases.AddItemToScheduleInput{
			Discipline:    item.Discipline,
			TeacherID:     item.TeacherID,
			StudentsCount: item.StudentsCount,
			Weekday:       item.Weekday,
			LessonNumber:  item.LessonNumber,
			Subgroup:      item.Subgroup,
			Weektype:      item.Weektype,
			LessonType:    item.LessonType,
			Classroom:     item.Classroom,
		}
	}

	err = h.schedule.AddItemsToSchedule(ctx, scheduleID, input)
	if err != nil {
		h.logger.Error("Add items to schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

type RemoveScheduleItemRequest struct {
	Weekday      *time.Weekday `json:"weekday"`
	LessonNumber int8          `json:"lesson_number"`
	Subgroup     int8          `json:"subgroup"`
	Weektype     *int8         `json:"weektype"`
}

// RemoveScheduleItem - DELETE /v1/schedules/:id/items
func (h *Handler) RemoveScheduleItem(c echo.Context) error {
	ctx := c.Request().Context()

	var rq []RemoveScheduleItemRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	input := make([]usecases.RemoveItemFromScheduleInput, len(rq))

	for i, item := range rq {
		input[i] = usecases.RemoveItemFromScheduleInput{
			Weekday:      item.Weekday,
			LessonNumber: item.LessonNumber,
			Subgroup:     item.Subgroup,
			Weektype:     item.Weektype,
		}
	}

	err = h.schedule.RemoveItemsFromSchedule(ctx, scheduleID, input)
	if err != nil {
		h.logger.Error("Remove items from schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

type ExportScheduleRequest struct {
	ScheduleID uuid.UUID `param:"id"`
	Format     string    `query:"format"`
	AsCalendar bool      `query:"as_calendar"`
}

// ExportSchedule - GET /v1/schedules/:id/export
func (h *Handler) ExportSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	var rq ExportScheduleRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	buffer := bytes.NewBuffer([]byte{})

	var exportErr error
	if rq.AsCalendar {
		exportErr = h.schedule.ExportCycledScheduleAsCalendar(ctx, rq.ScheduleID, rq.Format, buffer)
	} else {
		exportErr = h.schedule.ExportSchedule(ctx, rq.ScheduleID, rq.Format, buffer)
	}

	if exportErr != nil {
		h.logger.Error("Export schedule error", "error", exportErr)
		return exportErr
	}

	fname := fmt.Sprintf("%s-%s.csv", rq.ScheduleID, time.Now().Format("20060102150405"))

	return WrapResponse(http.StatusOK, buffer).SendAsFile(c, fname, rq.Format)
}

func scheduleDTOtoView(dto usecases.ScheduleDTO) Schedule {
	items := make([]ScheduleItem, 0, len(dto.Items))
	for _, item := range dto.Items {
		var wt *string
		if item.Weektype != nil {
			s := item.Weektype.String()
			wt = &s
		}

		items = append(items, ScheduleItem{
			Discipline:    item.Discipline,
			TeacherID:     item.TeacherID,
			Weekday:       item.Weekday.String(),
			StudentsCount: item.StudentsCount,
			Date:          item.Date,
			LessonNumber:  item.LessonNumber,
			Subgroup:      item.Subgroup,
			Weektype:      wt,
			Weeknum:       item.Weeknum,
			LessonType:    item.LessonType.String(),
			Classroom:     string(item.Classroom),
		})
	}

	return Schedule{
		ID:         dto.ID,
		EduGroupID: dto.EduGroupID,
		StartDate:  dto.StartDate,
		EndDate:    dto.EndDate,
		Items:      items,
	}
}
