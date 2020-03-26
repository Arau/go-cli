package openapi

import (
	"fmt"
	"net/http"

	"code.storageos.net/storageos/c2-cli/apiclient"
	"code.storageos.net/storageos/openapi"
)

// badRequestError indicates that the request made by the client is invalid.
type badRequestError struct {
	msg string
}

func (e badRequestError) Error() string {
	if e.msg == "" {
		return "bad request"
	}
	return e.msg
}

func newBadRequestError(msg string) badRequestError {
	return badRequestError{
		msg: msg,
	}
}

// notFoundError indicates that a resource involved in carrying out the API
// request was not found.
type notFoundError struct {
	msg string
}

func (e notFoundError) Error() string {
	if e.msg == "" {
		return "not found"
	}
	return e.msg
}

func newNotFoundError(msg string) notFoundError {
	return notFoundError{
		msg: msg,
	}
}

// conflictError indicates that the requested operation could not be carried
// out due to a conflict between the current state and the desired state.
type conflictError struct {
	msg string
}

func (e conflictError) Error() string {
	if e.msg == "" {
		return "conflict"
	}
	return e.msg
}

func newConflictError(msg string) conflictError {
	return conflictError{
		msg: msg,
	}
}

type openAPIError struct {
	inner openapi.Error
}

func (e openAPIError) Error() string {
	return e.inner.Error
}

func newOpenAPIError(err openapi.Error) openAPIError {
	return openAPIError{
		inner: err,
	}
}

// mapOpenAPIError will given err and its corresponding resp attempt to map the
// HTTP error to an application level error.
//
// err is returned as is when any of the following are true:
//
// 	 → resp is nil
// 	 → err is not a GenericOpenAPIError or the unexported openAPIError
//
// Some response codes must be mapped by the caller in order to provide useful
// application level errors:
//
//   → http.StatusBadRequest returns a badRequestError, which must have a 1-to-1
//   mapping to a context specific application error
//   → http.StatusNotFound returns a notFoundError, which must have a 1-to-1
//   mapping to a context specific application error
//   → http.StatusConflict returns a conflictError which must have a 1-to-1
//   mapping to a context specific application error
//
func mapOpenAPIError(err error, resp *http.Response) error {
	if resp == nil {
		return err
	}

	var details string
	switch v := err.(type) {
	case openapi.GenericOpenAPIError:
		switch model := v.Model().(type) {
		case openapi.Error:
			details = model.Error
		default:
			details = fmt.Sprintf("%s", v.Body())
		}
	case openAPIError:
		details = v.Error()
	default:
		return err
	}

	switch resp.StatusCode {

	// 4XX
	case http.StatusBadRequest:
		return newBadRequestError(details)

	case http.StatusUnauthorized:
		return apiclient.NewAuthenticationError(details)

	case http.StatusForbidden:
		return apiclient.NewUnauthorisedError(details)

	case http.StatusNotFound:
		return newNotFoundError(details)

	case http.StatusConflict:
		return newConflictError(details)

	case http.StatusPreconditionFailed:
		return apiclient.NewStaleWriteError(details)

	case http.StatusUnprocessableEntity:
		return apiclient.NewInvalidStateTransitionError(details)

	// TODO(CP-3925): This may need changing to present a friendly error, or
	// it may be done up the call stack.
	case http.StatusUnavailableForLegalReasons:
		return apiclient.NewLicenceCapabilityError(details)

	// 5XX
	case http.StatusInternalServerError:
		return apiclient.NewServerError(details)

	case http.StatusServiceUnavailable:
		return apiclient.NewStoreError(details)

	default:
		return err
	}
}
