package errkit

// Error is string as error
type Error string

func (e Error) Error() string {
	return string(e)
}
