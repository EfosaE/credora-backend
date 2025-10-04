package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func RegisterOpenAPIRoutes(api chi.Router) {
	// 📌 Serve raw OpenAPI YAML
	api.Get("/documentation/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./internal/docs/openapi.yml")
	})

	// 📌 Serve Swagger UI (inline HTML)
	api.Get("/documentation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `
        <!DOCTYPE html>
        <html>
          <head>
            <title>Credora API Docs</title>
            <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
          </head>
          <body>
            <div id="swagger-ui"></div>
            <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
            <script>
              window.onload = () => {
                SwaggerUIBundle({
                  url: '/api/v1/documentation/openapi.yml',
                  dom_id: '#swagger-ui'
                });
              };
            </script>
          </body>
        </html>`
		w.Write([]byte(html))
	})

	// 📌 Serve static Swagger UI (optional, if you copy dist files)
	fs := http.FileServer(http.Dir("./public/swagger"))
	api.Handle("/documentation/*", http.StripPrefix("/documentation", fs))
	// r.Route("/reserved-account", func(r chi.Router) {
	// 	r.Delete("/{acctRef}", h.DeleteCustomerHandler)
	// })
}
