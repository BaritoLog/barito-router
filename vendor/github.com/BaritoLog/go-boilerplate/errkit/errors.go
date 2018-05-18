package errkit

import "bytes"

// Errors is array of error as error itself
type Errors []error

// Errors is return as string
func (e Errors) String(sep string) (str string) {
	var buffer bytes.Buffer

	for i, err := range e {
		if i > 0 {
			buffer.WriteString(sep)
		}
		buffer.WriteString(err.Error())
	}

	str = buffer.String()
	return
}

// Error return error message
func (e Errors) Error() (str string) {
	return e.String(": ")
}
