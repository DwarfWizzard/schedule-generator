package edugroups

import (
	"errors"
	"schedule-generator/internal/common"
	eduplans "schedule-generator/internal/domain/edu_plans"
	"time"

	"github.com/google/uuid"
)

const (
	EduactionStartMonth = time.September
	EducationStartDay   = 1
)

type EduGroup struct {
	ID            uuid.UUID
	Number        string
	EduPlanID     uuid.UUID
	Profile       string
	AdmissionYear int64
}

func NewEduGroup(number string, eduPlan *eduplans.EduPlan) (*EduGroup, error) {
	if len(number) == 0 {
		return nil, errors.New("invalid group number")
	}

	if eduPlan == nil {
		return nil, errors.New("edu plan can not be empty")
	}

	return &EduGroup{
		ID:            uuid.New(),
		Number:        number,
		Profile:       eduPlan.Profile,
		AdmissionYear: eduPlan.Year,
		EduPlanID:     eduPlan.ID,
	}, nil
}

// GetEducationStartDate start of current education year by semester number
func (e EduGroup) GetEducationStartDateBySemester(semester int) time.Time {
	if semester < 0 {
		semester = 1
	} else if semester%2 == 0 {
		semester -= 1
	}

	course := semester / 2

	return time.Date(int(e.AdmissionYear)+course, EduactionStartMonth, EducationStartDay, 0, 0, 0, 0, common.DefaultTimezone)
}

func (e *EduGroup) Validate() error {
	if len(e.Number) == 0 {
		return errors.New("invalid group number")
	}

	return nil
}
