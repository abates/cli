package cli

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrUsage = errors.New("Invalid Usage")

	ErrUnknownCommand  = fmt.Errorf("%w Unknown command", ErrUsage)
	ErrRequiredCommand = fmt.Errorf("%w A command is required", ErrUsage)
	ErrNoCommandFunc   = fmt.Errorf("%w No callback function was provided", ErrUsage)

	errParse        = errors.New("parse error")
	errRange        = errors.New("value out of range")
	errNumArguments = fmt.Errorf("%w not enough arguments given", ErrUsage)
)

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errRange
	}
	return ne.Err
}
