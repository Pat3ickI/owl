package owl

// ErrorType defines types of errors that are possible from soup
type ErrorType int

const (
	// ErrUnableToParse will be returned when the HTML could not be parsed
	ErrUnableToParse ErrorType = iota
	// ErrElementNotFound will be returned when element was not found
	ErrElementNotFound
	// Just like ErrElementNotFound But this deal with Plural forms
	ErrElementsNotFound
	// ErrNoNextSibling will be returned when no next sibling can be found
	ErrNoNextSibling
	// ErrNoPreviousSibling will be returned when no previous sibling can be found
	ErrNoPreviousSibling
	// ErrNoNextElementSibling will be returned when no next element sibling can be found
	ErrNoNextElementSibling
	// ErrNoPreviousElementSibling will be returned when no previous element sibling can be found
	ErrNoPreviousElementSibling
	// ErrCreatingGetRequest will be returned when the get request couldn't be created
	ErrCreatingGetRequest
	// ErrInGetRequest will be returned when there was an error during the get request
	ErrInGetRequest
	// ErrCreatingPostRequest will be returned when the post request couldn't be created
	ErrCreatingPostRequest
	// ErrMarshallingPostRequest will be returned when the body of a post request couldn't be serialized
	ErrMarshallingPostRequest
	// ErrReadingResponse will be returned if there was an error reading the response to our get request
	ErrReadingResponse
)

// Error allows easier introspection on the type of error returned.
// If you know you have a Error, you can compare the Type to one of the exported types
// from this package to see what kind of error it is, then further inspect the Error() method
// to see if it has more specific details for you, like in the case of a ErrElementNotFound
// type of error.
type Error struct {
	Type ErrorType
	msg  error
}

func (er *Error) Err() error {
	return er.msg
}

func newError(t ErrorType, msg error) *Error {
	return &Error{Type: t, msg: msg}
}

// type Error struct {
// 	Type ErrorType
// 	msg  string
// }

// func (se Error) Error() string {
// 	return se.msg
// }

// func newError(t ErrorType, msg string) Error {
// 	return Error{Type: t, msg: msg}
// }
