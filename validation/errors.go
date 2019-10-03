package validation

import "fmt"

type ValidationError struct {
	Message       string
	RelevantLines []LineContext
}

func NewError(format string, args ...interface{}) *ValidationError {
	return &ValidationError{
		Message: fmt.Sprintf(format, args...),
	}
}

func (ve *ValidationError) At(lineText string, lineIndex int) *ValidationError {
	return ve.Link(LineContext{
		LineText:  lineText,
		LineIndex: lineIndex,
	})
}

func (ve *ValidationError) Link(lineContexts ...LineContext) *ValidationError {
	ve.RelevantLines= append(ve.RelevantLines, lineContexts...)
	return ve
}

func (v *Validator) Errors() []*ValidationError {
	return v.errors[:]
}

func (v *Validator) addError(format string, args ...interface{}) *ValidationError {
	err := NewError(format, args...)
	v.errors = append(v.errors,err)
	return err
}
