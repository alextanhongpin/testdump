package grpcdump

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Status is a struct that represents the status of a gRPC call.
type Status struct {
	Code    string     `json:"code"`    // Code is the string representation of the status code.
	Number  codes.Code `json:"number"`  // Number is the numerical representation of the status code.
	Message string     `json:"message"` // Message is the error message associated with the status code.
}

// newStatus is a function that takes an error and returns a new Status object.
// If the error can't be converted to a gRPC status, it returns nil.
func newStatus(err error) *Status {
	sts, ok := status.FromError(err)
	if !ok {
		return nil // Return nil if error is not a gRPC status error.
	}

	// Return a new Status object with the code, number, and message from the gRPC status.
	return &Status{
		Code:    sts.Code().String(), // Convert the status code to a string.
		Number:  sts.Code(),          // Get the numerical representation of the status code.
		Message: sts.Message(),       // Get the error message associated with the status code.
	}
}
