package middleware

import (
	"net/http"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/response"
)

// BodyLimit limits the maximum request body size.
// Requests exceeding the limit receive a 413 Payload Too Large response.
//
//	mux.Handle("/upload", middleware.BodyLimit(10<<20)(handler)) // 10 MB
func BodyLimit(maxBytes int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				response.Err(w, errors.New(errors.CodeRequestTooLarge,
					"Request body too large").WithStatus(http.StatusRequestEntityTooLarge))
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
