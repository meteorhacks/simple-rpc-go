package srpc

// Error method makes the Error type implement the error interface
func (e *Error) Error() (message string) {
	return e.Message
}
