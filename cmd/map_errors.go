package cmd

import (
	"context"
	"errors"

	"code.storageos.net/storageos/c2-cli/apiclient"
)

// ErrCommandTimedOut is returned when a command's execution deadline is
// exceeded.
var ErrCommandTimedOut = errors.New("timed out performing command")

// MapCommandError attempts to map err to a user friendly error type. If
// err is not a known application-level error mapping it is returned as
// is.
func MapCommandError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return ErrCommandTimedOut
	default:
		return err
	}
}

// ExitCodeForError returns the appropriate application exit code for err.
func ExitCodeForError(err error) int {
	switch {
	case errors.Is(err, apiclient.BadRequestError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.AuthenticationError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.UnauthorisedError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.NotFoundError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.ConflictError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.StaleWriteError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.InvalidStateTransitionError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.LicenceCapabilityError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.ServerError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, apiclient.StoreError{}):
		return 1 // TODO: Pick code
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, ErrCommandTimedOut):
		return 124
	default:
		return 1
	}
}
