package request

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/KARTIKrocks/apikit/errors"
)

// PathParam returns a path parameter value from the request.
// Works with Go 1.22+ stdlib routing: http.ServeMux patterns like "GET /users/{id}"
//
//	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
//	    id := request.PathParam(r, "id")
//	})
func PathParam(r *http.Request, key string) string {
	return r.PathValue(key)
}

// PathParamRequired returns a path parameter or an error if empty.
func PathParamRequired(r *http.Request, key string) (string, error) {
	val := r.PathValue(key)
	if val == "" {
		return "", errors.BadRequest(fmt.Sprintf("Path parameter %q is required", key))
	}
	return val, nil
}

// PathParamInt returns a path parameter as an int.
func PathParamInt(r *http.Request, key string) (int, error) {
	val := r.PathValue(key)
	if val == "" {
		return 0, errors.BadRequest(fmt.Sprintf("Path parameter %q is required", key))
	}

	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.BadRequest(fmt.Sprintf("Path parameter %q must be an integer", key)).
			WithField(key, "must be an integer")
	}
	return n, nil
}

// PathParamInt64 returns a path parameter as an int64.
func PathParamInt64(r *http.Request, key string) (int64, error) {
	val := r.PathValue(key)
	if val == "" {
		return 0, errors.BadRequest(fmt.Sprintf("Path parameter %q is required", key))
	}

	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, errors.BadRequest(fmt.Sprintf("Path parameter %q must be an integer", key)).
			WithField(key, "must be an integer")
	}
	return n, nil
}
