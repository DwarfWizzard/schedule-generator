package departments

import (
	"errors"

	"github.com/google/uuid"
)

type Department struct {
	ID         uuid.UUID
	ExternalID string
	FacultyID  uuid.UUID
	Name       string
}

// NewDepartment
func NewDepartment(facultyID uuid.UUID, externalID, name string) (*Department, error) {
	if len(name) == 0 {
		return nil, errors.New("invalid name value")
	}

	if len(externalID) == 0 {
		return nil, errors.New("invalid external id value")
	}

	return &Department{
		ID:         uuid.New(),
		Name:       name,
		ExternalID: externalID,
		FacultyID:  facultyID,
	}, nil
}

func (d *Department) Validate() error {
	if len(d.Name) == 0 {
		return errors.New("invalid name value")
	}

	if len(d.ExternalID) == 0 {
		return errors.New("invalid external id value")
	}

	return nil
}
