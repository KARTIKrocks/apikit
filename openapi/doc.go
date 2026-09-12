// Package openapi generates an OpenAPI 3.1 document from apikit request/response
// types and a router.Router, with zero third-party dependencies.
//
// Request and response schemas are derived by reflecting on the same `json` and
// `validate` struct tags that request.Bind[T] and request.ValidateStruct already
// use, so a type only needs to be described once:
//
//	type CreateUserReq struct {
//	    Name  string `json:"name"  validate:"required,min=2,max=100"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	doc := openapi.New(openapi.Info{Title: "My API", Version: "1.0.0"})
//	doc.Add("POST", "/users", openapi.Operation{
//	    Summary: "Create a user",
//	    Request: CreateUserReq{},
//	    Responses: map[int]openapi.Response{
//	        201: {Description: "Created", Body: User{}},
//	    },
//	})
//
//	// Fill in any route the router knows about but that Add hasn't described yet.
//	doc.Sync(r)
//
//	r.GetFunc("/openapi.json", doc.Handler())
//	r.GetFunc("/docs", doc.SwaggerUIHandler("/openapi.json"))
//
// Cross-field and programmatic validation (request.NewValidation) is out of
// scope for schema generation, matching the tag engine's own documented limits.
package openapi
