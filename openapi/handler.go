package openapi

import (
	"fmt"
	"html"
	"net/http"
)

// swaggerUITemplate renders a minimal Swagger UI page that loads its assets
// (swagger-ui-dist) from jsdelivr and points them at specURL. This keeps the
// core module dependency-free — no Go package is involved — at the cost of
// requiring the browser to reach the CDN.
const swaggerUITemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>%s</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: %q,
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      });
    };
  </script>
</body>
</html>
`

// SwaggerUIHandler serves a Swagger UI page that loads the spec from
// specURL — typically the path where Handler is mounted, e.g. "/openapi.json".
func (d *Document) SwaggerUIHandler(specURL string) http.HandlerFunc {
	title := html.EscapeString(d.info.Title)
	page := fmt.Sprintf(swaggerUITemplate, title, specURL)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}
