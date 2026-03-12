package openapi

import (
	"strings"

	"github.com/go-chi/chi/v5"
)

var ignoredPrefixes = []string{
	"/openapi.json",
	"/documentation",
	"/metrics",
	"/reserved-account",
}

func shouldIgnore(path string) bool {
	for _, p := range ignoredPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// GenerateSpec builds OpenAPI spec from a chi.Router
func GenerateSpec(r *chi.Mux, info Info, servers []Server) OpenAPISpec {
	paths := buildPaths(r)

	return OpenAPISpec{
		OpenAPI: "3.0.3",
		Info:    info,
		Servers: servers,
		Paths:   paths,
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
					Format: "JWT",
				},
			},
			Schemas: schemaRegistry, // schemas registered via registerSchema
		},
	}
}

// walk chi routes and build paths
func buildPaths(r *chi.Mux) map[string]PathItem {
	paths := map[string]PathItem{}
	walkRoutes(r.Routes(), paths, "", "/api/v1")
	return paths
}

func walkRoutes(routes []chi.Route, paths map[string]PathItem, prefix string, basePath string) {
	for _, rt := range routes {
		currentPattern := prefix
		if rt.Pattern != "/" {
			currentPattern += rt.Pattern
		}

		// strip trailing /*
		clean := strings.ReplaceAll(currentPattern, "/*", "")

		if clean != basePath {
			relativePath := strings.Replace(clean, basePath, "", 1)

			if shouldIgnore(relativePath) {
				continue
			}

			pathItem := PathItem{}
			for method, handler := range rt.Handlers {
				if method == "*" {
					continue
				}
				_ = handler // no longer needed for lookup
				doc := getDoc(method, relativePath)
				op := buildOperation(relativePath, method, doc)
				pathItem[strings.ToLower(method)] = op
			}

			if len(pathItem) > 0 {
				paths[relativePath] = pathItem
			}
		}

		if rt.SubRoutes != nil {
			walkRoutes(rt.SubRoutes.Routes(), paths, currentPattern, basePath)
		}
	}
}

// buildOperation builds a single OpenAPI operation
func buildOperation(path, method string, doc *Doc) Operation {
	tag := inferTag(path)
	params := extractPathParams(path)
	protected := isProtectedRoute(path)

	var security []SecurityReq
	if protected {
		security = []SecurityReq{{"bearerAuth": []string{}}}
	}

	responses := map[string]Response{
		"400": {Description: "Bad Request"},
		"500": {Description: "Internal Server Error"},
	}
	if protected {
		responses["401"] = Response{Description: "Unauthorized"}
	}

	var summary, description string
	var requestBody *RequestBody

	if doc != nil {
		summary = doc.Summary
		description = doc.Description

		if doc.Request != nil {
			schemaName := registerSchema(doc.Request)
			requestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{Ref: "#/components/schemas/" + schemaName},
					},
				},
			}
		}

		if doc.Response != nil {
			schemaName := registerSchema(doc.Response)
			responses["200"] = Response{
				Description: "Success",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{Ref: "#/components/schemas/" + schemaName},
					},
				},
			}
		} else {
			responses["200"] = Response{Description: "Success"}
		}

		// -- Header params --
		for _, h := range doc.Headers {
			params = append(params, Parameter{
				Name:        h.Name,
				In:          "header",
				Required:    h.Required,
				Description: h.Description,
				Schema:      Schema{Type: "string"},
			})
		}
	} else {
		responses["200"] = Response{Description: "Success"}
	}

	return Operation{
		Summary:     summary,
		Description: description,
		Tags:        []string{tag},
		Parameters:  params,
		RequestBody: requestBody,
		Responses:   responses,
		Security:    security,
	}
}

func inferTag(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "default"
}

func extractPathParams(path string) []Parameter {
	var params []Parameter
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := strings.Trim(part, "{}")
			params = append(params, Parameter{
				Name:     name,
				In:       "path",
				Required: true,
				Schema:   Schema{Type: "string"},
			})
		}
	}
	return params
}

var publicPrefixes = []string{
	"/auth/",
	"/health/",
	"/webhooks/",
	"/simulator/",
	"/tests/",
}

func isProtectedRoute(path string) bool {
	for _, prefix := range publicPrefixes {
		if strings.Contains(path, prefix) {
			return false
		}
	}
	return true
}


	// if doc.Response != nil {
	// 		schemaName := registerSchema(doc.Response)
	// 		envelopeSchema := buildEnvelopeSchema(schemaName)

	// 		// register the envelope itself as a named schema
	// 		envelopeName := schemaName + "Response"
	// 		schemaRegistry[envelopeName] = envelopeSchema

	// 		responses["200"] = Response{
	// 			Description: "Success",
	// 			Content: map[string]MediaType{
	// 				"application/json": {
	// 					Schema: Schema{Ref: "#/components/schemas/" + envelopeName},
	// 				},
	// 			},
	// 		}
	// 	}