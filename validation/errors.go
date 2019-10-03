package validation

type ValidationError struct {
	Message   string
	LineText  string
	LineIndex int
}

func (v *Validator) Errors() []ValidationError {
	return v.errors[:]
}

//
// Helpers

func (v *Validator) addError(err ValidationError) {
	v.errors = append(v.errors, err)
}

func (v *Validator) addLineError(line string, message string) {
	v.addError(ValidationError{
		Message:   message,
		LineText:  line,
		LineIndex: v.lines,
	})
}
