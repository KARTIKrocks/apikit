// Command openapi demonstrates generating and serving an OpenAPI 3.1 document
// straight from apikit request/response types and a router.Router.
//
// Run it and visit:
//   - http://localhost:8080/openapi.json — the generated spec
//   - http://localhost:8080/docs         — Swagger UI, pointed at the spec above
package main

import (
	"log"
	"net/http"

	"github.com/KARTIKrocks/apikit/openapi"
	"github.com/KARTIKrocks/apikit/request"
	"github.com/KARTIKrocks/apikit/response"
	"github.com/KARTIKrocks/apikit/router"
)

type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"omitempty,oneof=admin user mod"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func main() {
	r := router.New()

	r.Post("/users", func(w http.ResponseWriter, req *http.Request) error {
		in, err := request.Bind[CreateUserReq](req)
		if err != nil {
			return err
		}
		response.Created(w, "User created", User{ID: "1", Name: in.Name, Email: in.Email, Role: in.Role})
		return nil
	})
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) error {
		response.OK(w, "OK", User{ID: request.PathParam(req, "id"), Name: "Alice"})
		return nil
	})

	doc := openapi.New(openapi.Info{
		Title:       "Example API",
		Version:     "1.0.0",
		Description: "Generated from apikit request/response types.",
	})
	doc.Add("POST", "/users", openapi.Operation{
		Summary: "Create a user",
		Tags:    []string{"users"},
		Request: CreateUserReq{},
		Responses: map[int]openapi.Response{
			201: {Description: "User created", Body: User{}},
			422: {Description: "Validation failed"},
		},
	})
	doc.Add("GET", "/users/{id}", openapi.Operation{
		Summary: "Get a user by ID",
		Tags:    []string{"users"},
		Responses: map[int]openapi.Response{
			200: {Description: "The user", Body: User{}},
			404: {Description: "User not found"},
		},
	})
	// Picks up any other route registered on r that Add hasn't described yet,
	// so the spec never silently falls out of sync with the router.
	doc.Sync(r)

	r.GetFunc("/openapi.json", doc.Handler())
	r.GetFunc("/docs", doc.SwaggerUIHandler("/openapi.json"))

	log.Println("listening on :8080 — try /docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}
