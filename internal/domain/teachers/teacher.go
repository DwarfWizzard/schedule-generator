package teachers

import (
	"errors"

	"github.com/google/uuid"
)

// TODO: make value-object for position and degree
type Teacher struct {
	ID           uuid.UUID
	ExternalID   string
	Name         string
	Position     string
	Degree       string
	DepartmentID uuid.UUID
}

// NewTeacher
func NewTeacher(departmentID uuid.UUID, externalID, name, position, degree string) (*Teacher, error) {
	if len(name) == 0 {
		return nil, errors.New("invalid name value")
	}

	if len(externalID) == 0 {
		return nil, errors.New("invalid external id value")
	}

	if len(position) == 0 {
		return nil, errors.New("invalid position value")
	}

	if len(degree) == 0 {
		return nil, errors.New("invalid degree value")
	}

	return &Teacher{
		ID:           uuid.New(),
		ExternalID:   externalID,
		Name:         name,
		Position:     position,
		Degree:       degree,
		DepartmentID: departmentID,
	}, nil
}

func (t *Teacher) Validate() error {
	if len(t.Name) == 0 {
		return errors.New("invalid name value")
	}

	if len(t.ExternalID) == 0 {
		return errors.New("invalid external id value")
	}

	if len(t.Position) == 0 {
		return errors.New("invalid position value")
	}

	if len(t.Degree) == 0 {
		return errors.New("invalid degree value")
	}

	return nil
}
