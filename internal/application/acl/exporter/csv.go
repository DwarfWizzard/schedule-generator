package exporter

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"schedule-generator/internal/domain/schedules"
	"strconv"
)

var cycledCsvHeader = []string{
	"Group",
	"Day",
	"Les",
	"Aud",
	"Week",
	"Subg",
	"Name",
	"Caf",
	"Subject",
	"Subj_Type",
	"Start",
	"End",
	"Subj_CafID",
	"PrepID",
}

var calendarCsvHeader = []string{
	"Group",
	"StudInLesson",
	"Day",
	"Les",
	"Aud",
	"Week",
	"Subg",
	"Name",
	"CafID",
	"Subject",
	"Subj_Type",
	"Date",
	"Subj_CafID",
	"PrepID",
	"Themas",
	"Substitution_Name",       // leaved empty
	"Substitution_PrepID",     // leaved empty
	"Substitution_Subject",    // leaved empty
	"Substitution_Subj_type",  // leaved empty
	"Substitution_Subj_CafID", // leaved empty
	"Lesson_ID",               // leaved empty
	"Lesson_Num",              // leaved empty
}

type scheduleItemHandler func(ctx context.Context, groupNumber string, item schedules.ScheduleItem) ([]string, error)

type csvExporter struct {
	repo   ExporterRepository
	logger *slog.Logger
}

func (exp *csvExporter) Export(ctx context.Context, schedule *schedules.Schedule, dst io.Writer) error {
	var header []string
	var handler scheduleItemHandler
	var listItems []schedules.ScheduleItem

	switch schedule.Type {
	case schedules.ScheduleTypeCycled:
		header = cycledCsvHeader
		handler = exp.cycledScheduleItemHandler
		listItems = schedule.Cycled.ListItem()

	case schedules.ScheduleTypeCalendar:
		header = calendarCsvHeader
		handler = exp.calendarScheduleItemHandler
		listItems = schedule.Calendar.ListItem()

	default:
		return errors.New("unsupported schedule type")
	}

	logger := exp.logger.With("schedule_id", schedule.ID)

	group, err := exp.repo.GetEduGroup(ctx, schedule.EduGroupID)
	if err != nil {
		logger.Error("Get edu group error", "error", err)
		return err
	}

	stream := csv.NewWriter(dst)
	stream.Write(header)

	for _, item := range listItems {
		row, err := handler(ctx, group.Number, item)
		if err != nil {
			logger.Error("Handler schedule item error", "error", err)
			return err
		}

		stream.Write(row)
	}

	stream.Flush()
	return nil
}

func (exp *csvExporter) cycledScheduleItemHandler(ctx context.Context, groupNumber string, item schedules.ScheduleItem) ([]string, error) {
	var weekType string

	if item.Weektype != nil {
		switch *item.Weektype {
		case schedules.WeekTypeEven:
			weekType = "Ч"
		case schedules.WeekTypeUneven:
			weekType = "Н"
		default:
			//leave weektype empty
		}
	}

	var subgroup string
	if item.Subgroup > 0 {
		subgroup = strconv.FormatInt(int64(item.Subgroup), 10)
	}

	teacher, err := exp.repo.GetTeacher(ctx, item.TeacherID)
	if err != nil {
		return nil, fmt.Errorf("get teacher error: %w", err)
	}

	department, err := exp.repo.GetDepartment(ctx, teacher.DepartmentID)
	if err != nil {
		return nil, fmt.Errorf("get department error: %w", err)
	}

	lessonType, err := formLessonType(item.LessonType)
	if err != nil {
		return nil, err
	}

	return []string{
		groupNumber,
		strconv.FormatInt(int64(item.Weekday), 10),
		strconv.FormatInt(int64(item.LessonNumber)+1, 10),
		formCabinetAddress(item.Cabinet),
		weekType,
		subgroup,
		teacher.Name,
		department.ExternalID,
		item.Discipline,
		lessonType,
		"-100",
		"-100",
		"0",
		teacher.ExternalID,
	}, nil
}

func (exp *csvExporter) calendarScheduleItemHandler(ctx context.Context, groupNumber string, item schedules.ScheduleItem) ([]string, error) {
	var subgroup string
	if item.Subgroup > 0 {
		subgroup = strconv.FormatInt(int64(item.Subgroup), 10)
	}

	teacher, err := exp.repo.GetTeacher(ctx, item.TeacherID)
	if err != nil {
		return nil, fmt.Errorf("get teacher error: %w", err)
	}

	department, err := exp.repo.GetDepartment(ctx, teacher.DepartmentID)
	if err != nil {
		return nil, fmt.Errorf("get department error: %w", err)
	}

	lessonType, err := formLessonType(item.LessonType)
	if err != nil {
		return nil, err
	}

	weeknum := "0"
	if item.Weeknum != nil {
		weeknum = strconv.FormatInt(int64(*item.Weeknum), 10)
	}

	return []string{
		groupNumber,
		strconv.FormatInt(int64(item.StudentsCount), 10),
		strconv.FormatInt(int64(item.Weekday), 10),
		strconv.FormatInt(int64(item.LessonNumber)+1, 10),
		formCabinetAddress(item.Cabinet),
		weeknum,
		subgroup,
		teacher.Name,
		department.ExternalID,
		item.Discipline,
		lessonType,
		item.Date.Format("02.01.2006"),
		"",
		teacher.ExternalID,
		"",
	}, nil
}

func formLessonType(lessonType schedules.ItemLessonType) (string, error) {
	switch lessonType {
	case schedules.ItemTypeLecture:
		return "лек.", nil
	case schedules.ItemTypePractice:
		return "пр.", nil
	case schedules.ItemTypeSeminar:
		return "сем.", nil
	case schedules.ItemTypeExam:
		return "экз.", nil
	case schedules.ItemTypeLaboratory:
		return "лаб.", nil
	default:
		return "", fmt.Errorf("unknown lesson type %s", lessonType.String())
	}
}

func formCabinetAddress(cabinet schedules.Cabinet) string {
	return fmt.Sprintf("%s-%s", cabinet.Building, cabinet.Auditorium)
}
