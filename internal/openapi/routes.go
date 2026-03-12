// internal/openapi/routes.go

package openapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterOpenAPIRoutes(r chi.Router) {
	// specBytes, _ := json.MarshalIndent(spec, "", "  ")

	// // Serve the raw JSON spec
	// r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write(specBytes)
	// })

	// Serve SwaggerUI (via CDN — no npm needed)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Credora API Docs</title>
    <meta charset="utf-8"/>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
SwaggerUIBundle({
    url: "/api/v1/openapi.json",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout",
    persistAuthorization: true,
})
</script>
</body>
</html>`
		w.Write([]byte(html))
	})
}
