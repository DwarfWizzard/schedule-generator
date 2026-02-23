package eduplans

import (
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

type Module struct {
	Discipline string
}

type EduPlan struct {
	ID           uuid.UUID
	DirectionID  uuid.UUID
	DepartmentID uuid.UUID
	Profile      string
	Year         int64

	Modules []Module
}

func (p *EduPlan) Validate() error {
	if p.Year < 1900 || p.Year > 2999 {
		return errors.New("invalid year value")
	}

	if len(p.Profile) == 0 {
		return errors.New("invalid profile")
	}

	return nil
}

// NewEduPlan
func NewEduPlan(directionID, departmentID uuid.UUID, profile string, year int64) (*EduPlan, error) {
	p := &EduPlan{
		ID:           uuid.New(),
		DirectionID:  directionID,
		DepartmentID: departmentID,
		Profile:      profile,
		Year:         year,
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

var errModuleNotFound = errors.New("module not found")

// GetModule returns module from education plan by discipline
func (e *EduPlan) GetModule(discipline string) (*Module, error) {
	idx := slices.IndexFunc(e.Modules, func(m Module) bool {
		return m.Discipline == discipline
	})
	if idx < 0 {
		return nil, errModuleNotFound
	}

	return &e.Modules[idx], nil
}

// ListModule returns module list
func (e *EduPlan) ListModule() []Module {
	return e.Modules
}

// AddModule
func (e *EduPlan) AddModule(
	discipline string,
) (*Module, error) {
	if e == nil {
		return nil, nil
	}

	m, _ := e.GetModule(discipline)
	if m != nil {
		return nil, fmt.Errorf("module for discipline %s already exists", discipline)
	}

	module := Module{
		Discipline: discipline,
	}

	e.Modules = append(e.Modules, module)

	return &module, nil
}
