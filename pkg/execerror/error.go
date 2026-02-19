package execerror

import (
	"strings"
)

type ExecErrorType = string

const (
	TypeInternal           = "Internal Error"
	TypeInvalidInput       = "Invalid Input"
	TypeProcessingConflict = "Processing Conflict"
	TypeUnimpemented       = "Unimplemented"
	TypeForbbiden          = "Forbidden"
)

type ExecError struct {
	Type    ExecErrorType
	Cause   error
	Details map[string][]string
}

func NewExecError(errType ExecErrorType, cause error) *ExecError {
	return &ExecError{
		Type:  errType,
		Cause: cause,
	}
}

func (e ExecError) Error() string {
	var b strings.Builder
	b.WriteString(e.Type)

	if e.Cause != nil {
		b.WriteRune(',')
		b.WriteString(e.Cause.Error())
	}

	if len(e.Details) > 0 {
		b.WriteRune(',')

		for k, d := range e.Details {
			b.WriteString(k)
			b.WriteRune(':')
			b.WriteString(strings.Join(d, ", "))
			b.WriteRune('\t')
		}
	}

	return b.String()
}

func (e *ExecError) AddDetails(name string, data ...string) *ExecError {
	if e == nil {
		return nil
	}

	if e.Details == nil {
		e.Details = make(map[string][]string)
	}

	if _, ok := e.Details[name]; !ok {
		e.Details[name] = make([]string, 0, len(data))
	}

	e.Details[name] = append(e.Details[name], data...)
	return e
}
