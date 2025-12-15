package main

import (
	stderrors "errors"
)


// For reference, another way to do errors:
// // 1. Define a struct for your custom error type.
// type CustomError struct {
// 	Code int
// 	Message string
// }

// // 2. Implement the 'error' interface by providing the Error() string method.
// func (e *CustomError) Error() string {
// 	// This method defines the string representation of the error.
// 	return fmt.Sprintf("Error Code %d: %s", e.Code, e.Message)
// }


var (
	ErrOnWrite = stderrors.New("Could not write")
	KeyNotFound =  stderrors.New("Key Not found")
)