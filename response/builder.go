package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// Builder provides a fluent interface for constructing responses.
//
//	response.New().
//	    Status(http.StatusCreated).
//	    Message("User created").
//	    Data(user).
//	    Header("X-Resource-ID", user.ID).
//	    Send(w)
type Builder struct {
	statusCode int
	headers    map[string]string
	envelope   Envelope
}

// New creates a new response Builder with defaults (200 OK, success=true).
func New() *Builder {
	return &Builder{
		statusCode: http.StatusOK,
		envelope: Envelope{
			Success: true,
		},
	}
}

// Status sets the HTTP status code.
func (b *Builder) Status(code int) *Builder {
	b.statusCode = code
	b.envelope.Success = code >= 200 && code < 300
	return b
}

// Message sets the response message.
func (b *Builder) Message(msg string) *Builder {
	b.envelope.Message = msg
	return b
}

// Data sets the response data.
func (b *Builder) Data(data any) *Builder {
	b.envelope.Data = data
	return b
}

// Meta sets the response metadata.
func (b *Builder) Meta(meta any) *Builder {
	b.envelope.Meta = meta
	return b
}

// Err sets error information, marking the response as failed.
func (b *Builder) Err(code, message string) *Builder {
	b.envelope.Success = false
	b.envelope.Error = &ErrorBody{
		Code:    code,
		Message: message,
	}
	return b
}

// ErrFields adds field-level error details.
func (b *Builder) ErrFields(fields map[string]string) *Builder {
	if b.envelope.Error == nil {
		b.envelope.Error = &ErrorBody{}
	}
	b.envelope.Error.Fields = fields
	return b
}

// ErrDetails adds arbitrary error metadata.
func (b *Builder) ErrDetails(details map[string]any) *Builder {
	if b.envelope.Error == nil {
		b.envelope.Error = &ErrorBody{}
	}
	b.envelope.Error.Details = details
	return b
}

// Header sets a custom response header.
func (b *Builder) Header(key, value string) *Builder {
	if b.headers == nil {
		b.headers = make(map[string]string)
	}
	b.headers[key] = value
	return b
}

// Pagination sets offset-based pagination metadata.
func (b *Builder) Pagination(page, perPage, total int) *Builder {
	b.envelope.Meta = NewPageMeta(page, perPage, total)
	return b
}

// CursorPagination sets cursor-based pagination metadata.
func (b *Builder) CursorPagination(nextCursor string, hasMore bool, total int) *Builder {
	b.envelope.Meta = &CursorMeta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      total,
	}
	return b
}

// Send writes the response to the http.ResponseWriter.
// The timestamp is set at send time to reflect the actual response time.
func (b *Builder) Send(w http.ResponseWriter) {
	b.envelope.Timestamp = time.Now().Unix()
	for k, v := range b.headers {
		w.Header().Set(k, v)
	}
	write(w, b.statusCode, b.envelope)
}

// JSON returns the response envelope as JSON bytes.
func (b *Builder) JSON() ([]byte, error) {
	b.envelope.Timestamp = time.Now().Unix()
	return json.Marshal(b.envelope)
}

// MustJSON returns the response as JSON bytes or panics.
func (b *Builder) MustJSON() []byte {
	data, err := b.JSON()
	if err != nil {
		panic(err)
	}
	return data
}
