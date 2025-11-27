package edudirections

import (
	"errors"

	"github.com/google/uuid"
)

type EduDirection struct {
	ID           uuid.UUID
	Name         string
	DepartmentID uuid.UUID
}

// NewEduDirection
func NewEduDirection(departmentID uuid.UUID, name string) (*EduDirection, error) {
	if len(name) == 0 {
		return nil, errors.New("invalid name value")
	}

	return &EduDirection{
		ID:           uuid.New(),
		Name:         name,
		DepartmentID: departmentID,
	}, nil
}

func (e *EduDirection) Validate() error {
	if len(e.Name) == 0 {
		return errors.New("invalid name value")
	}

	return nil
}
