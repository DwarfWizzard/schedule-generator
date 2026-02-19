package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/common"
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ScheduleUsecase interface {
	CreateSchedule(ctx context.Context, input usecases.CreateScheduleInput, user *users.User) (*usecases.CreateScheduleOutput, error)
	ListSchedule(ctx context.Context, user *users.User) (usecases.ListScheduleOutput, error)
	GetSchedule(ctx context.Context, scheduleID uuid.UUID, user *users.User) (*usecases.GetScheduleOutput, error)
	AddItemsToSchedule(ctx context.Context, scheduleID uuid.UUID, input []usecases.AddItemToScheduleInput, user *users.User) error
	UpdateItemInSchedule(ctx context.Context, scheduleID uuid.UUID, input usecases.AddItemToScheduleInput, user *users.User) error
	RemoveItemsFromSchedule(ctx context.Context, scheduleID uuid.UUID, input []usecases.RemoveItemFromScheduleInput, user *users.User) error
	ExportSchedule(ctx context.Context, scheduleID uuid.UUID, format string, dst io.Writer, user *users.User) error
	ExportCycledScheduleAsCalendar(ctx context.Context, scheduleID uuid.UUID, format string, dst io.Writer, user *users.User) error
	UpdateSchedule(ctx context.Context, input usecases.UpdateScheduleInput, user *users.User) (*usecases.UpdateScheduleOutput, error)
	DeleteSchedule(ctx context.Context, scheduleID uuid.UUID, user *users.User) error
}

type ScheduleItem struct {
	Discipline        string     `json:"discipline"`
	TeacherID         uuid.UUID  `json:"teacher_id"`
	TeacherName       string     `json:"teacher_name"`
	Weekday           string     `json:"weekday"`
	StudentsCount     int16      `json:"students_count"`
	Date              *time.Time `json:"date"`
	LessonNumber      int8       `json:"lesson_number"`
	Subgroup          int8       `json:"subgroup"`
	Weektype          *int8      `json:"weektype"`
	Weeknum           *int       `json:"weeknum"`
	LessonType        int8       `json:"lesson_type"`
	CabinetAuditorium string     `json:"cabinet_auditorium"`
	CabinetBuilding   string     `json:"cabinet_building"`
}

type Schedule struct {
	ID             uuid.UUID      `json:"id"`
	EduGroupID     uuid.UUID      `json:"edu_group_id"`
	EduGroupNumber string         `json:"edu_group_number"`
	Semester       int            `json:"semester"`
	Type           string         `json:"type"`
	StartDate      *string        `json:"start_date"`
	EndDate        *string        `json:"end_date"`
	Items          []ScheduleItem `json:"items"`
}

type CreateScheduleRequest struct {
	//TODO: add calendar
	EduGroupID uuid.UUID `json:"edu_group_id"`
	Semester   int       `json:"semester"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
}

type CreateScheduleResponse struct {
	Schedule
}

// CreateScheduleRequest - POST /v1/schedules
func (h *Handler) CreateSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

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

	out, err := h.schedule.CreateSchedule(ctx, usecases.CreateScheduleInput{
		EduGroupID: rq.EduGroupID,
		Semester:   rq.Semester,
		StartDate:  startDate,
		EndDate:    endDate,
	}, user)
	if err != nil {
		h.logger.Error("Create schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, scheduleDTOtoView(out.ScheduleDTO, out.EduGroupNumber)).Send(c)
}

// ListSchedule - GET /v1/schedules
func (h *Handler) ListSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	list, err := h.schedule.ListSchedule(ctx, user)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	result := make([]Schedule, len(list))

	for idx, d := range list {
		result[idx] = scheduleDTOtoView(d.ScheduleDTO, d.EduGroupNumber)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

// GetSchedule - GET /v1/schedules/:id
func (h *Handler) GetSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.schedule.GetSchedule(ctx, scheduleID, user)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, scheduleDTOtoView(out.ScheduleDTO, out.EduGroupNumber)).Send(c)
}

type AddScheduleItemRequest struct {
	Discipline    string        `json:"discipline"`
	TeacherID     uuid.UUID     `json:"teacher_id"`
	Weekday       *time.Weekday `json:"weekday"`
	StudentsCount int16         `json:"students_count"`
	LessonNumber  int8          `json:"lesson_number"`
	Subgroup      int8          `json:"subgroup"`
	Weektype      *int8         `json:"weektype"`
	LessonType    int8          `json:"lesson_type"`
	CabinetID     uuid.UUID     `json:"cabinet_id"`
}

// AddScheduleItem - POST /v1/schedules/:id/items
func (h *Handler) AddScheduleItem(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

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
			CabinetID:     item.CabinetID,
		}
	}

	err = h.schedule.AddItemsToSchedule(ctx, scheduleID, input, user)
	if err != nil {
		h.logger.Error("Add items to schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

// UpdateScheduleItem - PUT /v1/schedules/:id/items
func (h *Handler) UpdateScheduleItem(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	var rq *AddScheduleItemRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	input := usecases.AddItemToScheduleInput{
		Discipline:    rq.Discipline,
		TeacherID:     rq.TeacherID,
		StudentsCount: rq.StudentsCount,
		Weekday:       rq.Weekday,
		LessonNumber:  rq.LessonNumber,
		Subgroup:      rq.Subgroup,
		Weektype:      rq.Weektype,
		LessonType:    rq.LessonType,
		CabinetID:     rq.CabinetID,
	}

	err = h.schedule.UpdateItemInSchedule(ctx, scheduleID, input, user)
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

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

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

	err = h.schedule.RemoveItemsFromSchedule(ctx, scheduleID, input, user)
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

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	var rq ExportScheduleRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	buffer := bytes.NewBuffer([]byte{})

	var exportErr error
	if rq.AsCalendar {
		exportErr = h.schedule.ExportCycledScheduleAsCalendar(ctx, rq.ScheduleID, rq.Format, buffer, user)
	} else {
		exportErr = h.schedule.ExportSchedule(ctx, rq.ScheduleID, rq.Format, buffer, user)
	}

	if exportErr != nil {
		h.logger.Error("Export schedule error", "error", exportErr)
		return exportErr
	}

	fname := fmt.Sprintf("%s-%s.csv", rq.ScheduleID, time.Now().Format("20060102150405"))

	return WrapResponse(http.StatusOK, buffer).SendAsFile(c, fname, rq.Format)
}

// DeleteSchedule - DELETE /v1/schedules/:id
func (h *Handler) DeleteSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	if err := h.schedule.DeleteSchedule(ctx, scheduleID, user); err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

type UpdateScheduleRequest struct {
	Semester  *int    `json:"semester"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

// UpdateSchedule - PATCH /v1/schedules/:id
func (h *Handler) UpdateSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	var rq UpdateScheduleRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	var startDate, endDate *time.Time
	if rq.StartDate != nil {
		if v, err := time.ParseInLocation(time.DateOnly, *rq.StartDate, common.DefaultTimezone); err != nil {
			return ErrInvalidInput
		} else {
			startDate = &v
		}
	}

	if rq.EndDate != nil {
		if v, err := time.ParseInLocation(time.DateOnly, *rq.EndDate, common.DefaultTimezone); err != nil {
			return ErrInvalidInput
		} else {
			endDate = &v
		}

	}

	log.Println("AAAAAAAAAAAAAAAAAAA", startDate)

	out, err := h.schedule.UpdateSchedule(ctx, usecases.UpdateScheduleInput{
		ID:        scheduleID,
		Semester:  rq.Semester,
		StartDate: startDate,
		EndDate:   endDate,
	}, user)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, scheduleDTOtoView(out.ScheduleDTO, out.EduGroupNumber)).Send(c)
}

func scheduleDTOtoView(dto usecases.ScheduleDTO, eduGroupNumber string) Schedule {
	var items []ScheduleItem

	if len(dto.Items) > 0 {
		items = make([]ScheduleItem, 0, len(dto.Items))
		for _, item := range dto.Items {
			var wt *int8
			if item.Weektype != nil {
				s := int8(*item.Weektype)
				wt = &s
			}

			items = append(items, ScheduleItem{
				Discipline:        item.Discipline,
				TeacherID:         item.TeacherID,
				TeacherName:       item.TeacherName,
				Weekday:           item.Weekday.String(),
				StudentsCount:     item.StudentsCount,
				Date:              item.Date,
				LessonNumber:      item.LessonNumber,
				Subgroup:          item.Subgroup,
				Weektype:          wt,
				Weeknum:           item.Weeknum,
				LessonType:        int8(item.LessonType),
				CabinetAuditorium: item.Cabinet.Auditorium,
				CabinetBuilding:   item.Cabinet.Building,
			})
		}
	}

	var startDate, endDate *string
	if dto.StartDate != nil {
		d := dto.StartDate.In(common.DefaultTimezone).Format(time.DateOnly)
		startDate = &d
	}

	if dto.EndDate != nil {
		d := dto.EndDate.In(common.DefaultTimezone).Format(time.DateOnly)
		endDate = &d
	}

	return Schedule{
		ID:             dto.ID,
		EduGroupID:     dto.EduGroupID,
		EduGroupNumber: eduGroupNumber,
		Semester:       dto.Semester,
		Type:           dto.Type.String(),
		StartDate:      startDate,
		EndDate:        endDate,
		Items:          items,
	}
}
