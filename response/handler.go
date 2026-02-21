package response

import (
	stderrors "errors"
	"net/http"

	"github.com/KARTIKrocks/apikit/errors"
)

// HandlerFunc is an HTTP handler that returns an error.
// When the handler returns a non-nil error, it is automatically
// converted to a structured JSON response.
//
//	mux.HandleFunc("GET /users/{id}", response.Handle(getUser))
//
//	func getUser(w http.ResponseWriter, r *http.Request) error {
//	    id := request.PathParam(r, "id")
//	    user, err := db.FindUser(id)
//	    if err != nil {
//	        return errors.NotFound("User")
//	    }
//	    response.OK(w, "Success", user)
//	    return nil
//	}
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handle wraps a HandlerFunc into a standard http.HandlerFunc.
// If the handler returns an error:
//   - *errors.Error → sends the appropriate status code and error body
//   - Standard error → sends 500 with a generic message (original error NOT exposed)
func Handle(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			handleError(w, err)
		}
	}
}

// HandleWith wraps a HandlerFunc with a custom error handler.
func HandleWith(fn HandlerFunc, onError func(w http.ResponseWriter, r *http.Request, err error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			onError(w, r, err)
		}
	}
}

// Wrap wraps a standard http.HandlerFunc so that any error passed to
// Err or the error convenience helpers is automatically written as a
// structured JSON response. Use this when you prefer the standard
// handler signature with no return type.
//
//	mux.HandleFunc("GET /users/{id}", response.Wrap(getUser))
//
//	func getUser(w http.ResponseWriter, r *http.Request) {
//	    id := request.PathParam(r, "id")
//	    user, err := db.FindUser(id)
//	    if err != nil {
//	        response.Err(w, errors.NotFound("User"))
//	        return
//	    }
//	    response.OK(w, "Success", user)
//	}
//
// Wrap is a thin pass-through today, but it future-proofs your handlers
// for middleware features like response capture or panic recovery.
func Wrap(fn http.HandlerFunc) http.HandlerFunc {
	return fn
}

// handleError converts an error to an HTTP response.
// Uses errors.As to correctly handle wrapped errors (e.g., fmt.Errorf("...: %w", apiErr)).
func handleError(w http.ResponseWriter, err error) {
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		// Don't expose internal error messages to clients
		apiErr = &errors.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       errors.CodeInternal,
			Message:    "An internal error occurred",
			Err:        err,
		}
	}

	Err(w, apiErr)
}
