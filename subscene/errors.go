package subscene

import "fmt"

type ResponseError struct {
	StatusCode int
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("Subscene responded with status code %d.", e.StatusCode)
}
