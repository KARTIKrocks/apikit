import CodeBlock from '../components/CodeBlock';

export default function GettingStarted() {
  return (
    <section id="getting-started" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-4">Getting Started</h2>

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Installation</h3>
      <p className="text-text-muted mb-3">Requires <strong>Go 1.22+</strong>.</p>
      <CodeBlock lang="bash" code={`go get github.com/KARTIKrocks/apikit`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Quick Start</h3>
      <p className="text-text-muted mb-3">
        A minimal API with request binding, structured errors, and graceful shutdown:
      </p>
      <CodeBlock code={`package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/KARTIKrocks/apikit/errors"
    "github.com/KARTIKrocks/apikit/middleware"
    "github.com/KARTIKrocks/apikit/request"
    "github.com/KARTIKrocks/apikit/response"
    "github.com/KARTIKrocks/apikit/router"
    "github.com/KARTIKrocks/apikit/server"
)

type CreateUserReq struct {
    Name  string \`json:"name" validate:"required,min=2"\`
    Email string \`json:"email" validate:"required,email"\`
}

func createUser(w http.ResponseWriter, r *http.Request) error {
    req, err := request.Bind[CreateUserReq](r)
    if err != nil {
        return err // Automatically sends 400 with structured error
    }

    if req.Email == "taken@example.com" {
        return errors.Conflict("Email already in use")
    }

    user := map[string]string{"id": "123", "name": req.Name}
    response.Created(w, "User created", user)
    return nil
}

func main() {
    r := router.New()
    r.Use(
        middleware.RequestID(),
        middleware.Recover(),
        middleware.Timeout(30 * time.Second),
    )

    r.Post("/users", createUser)

    srv := server.New(r, server.WithAddr(":8080"))
    srv.OnShutdown(func(ctx context.Context) error {
        log.Println("shutting down...")
        return nil
    })
    if err := srv.Start(); err != nil {
        log.Fatal(err)
    }
}`} />
    </section>
  );
}
