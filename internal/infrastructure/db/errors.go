package db

import "errors"

var (
	ErrorNotFound             = errors.New("not found")
	ErrorUniqueViolation      = errors.New("unique violation")
	ErrorAssociationViolation = errors.New("association violation")
)
