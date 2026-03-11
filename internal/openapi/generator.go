// internal/openapi/generator.go

package openapi

import (
	"strings"

	"github.com/go-chi/chi/v5"
)

func GenerateSpec(r *chi.Mux, info Info, servers []Server) OpenAPISpec {
	paths := buildPaths(r)

	return OpenAPISpec{
		OpenAPI: "3.0.3",
		Info:    info,
		Servers: servers,
		Paths:   paths,
		// Components holds reusable definitions.
		// Here we define what "bearerAuth" means — routes reference it by this name.
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
					Format: "JWT",
				},
			},
			Schemas: DefineSchemas(),
		},
	}
}

var ignoredPrefixes = []string{
	"/openapi.json",
	"/documentation",
	"/metrics",
}

func shouldIgnore(path string) bool {
	for _, p := range ignoredPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func buildPaths(r *chi.Mux) map[string]PathItem {
	paths := map[string]PathItem{}
	walkRoutes(r.Routes(), paths, "", "/api/v1")
	return paths
}

func walkRoutes(routes []chi.Route, paths map[string]PathItem, prefix string, basePath string) {
	for _, rt := range routes {
		currentPattern := prefix
		if rt.Pattern != "/" {
			currentPattern = prefix + rt.Pattern
		}

		// strip trailing /*
		clean := strings.ReplaceAll(currentPattern, "/*", "")

		if clean != basePath {
			// strip the basePath prefix so paths are relative
			relativePath := strings.Replace(clean, basePath, "", 1)
			// convert Chi {param} to OpenAPI {param} — already compatible!

			// skip ignored paths
			if shouldIgnore(relativePath) {
				continue
			}

			pathItem := PathItem{}
			for method := range rt.Handlers {
				if method == "*" {
					continue
				}
				op := buildOperation(relativePath, method)
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

// var responseSchemas = map[string]string{
// 	"get:/users/info":               "User",
// 	"get:/users/{email}":            "User",
// 	"get:/users/transactions":       "Transaction",
// 	"get:/transfers/{trxID}/status": "Transaction",
// }

// Auto-derive tags from path, extract path params, infer security
func buildOperation(path string, method string) Operation {
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

	// look up metadata for this specific route
	// key := method + ":" + path
	// meta, hasMeta := routeMetadata[key]
	// fix: normalize method to lowercase before lookup
	key := strings.ToLower(method) + ":" + path
	meta, hasMeta := routeMetadata[key]

	// attach response schema if defined
	if hasMeta && meta.SchemaName != "" {
		responses["200"] = Response{
			Description: "Success",
			Content: map[string]MediaType{
				"application/json": {
					Schema: Schema{Ref: "#/components/schemas/" + meta.SchemaName},
				},
			},
		}
	} else {
		responses["200"] = Response{Description: "Success"}
	}

	return Operation{
		Summary:     meta.Summary, // empty string if no metadata — omitempty handles it
		Description: meta.Description,
		Tags:        []string{tag},
		Parameters:  params,
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
	parts := strings.SplitSeq(path, "/")
	for part := range parts {
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

// publicPrefixes lists route segments that don't require a JWT.
// Any path NOT matching these gets the bearerAuth security requirement.
var publicPrefixes = []string{
	"/auth/",
	"/health/",
	"/webhooks/",
	"/simulator/",
	"/tests/",
}

// Routes that aren't public — customize to match your router logic
func isProtectedRoute(path string) bool {
	for _, prefix := range publicPrefixes {
		if strings.Contains(path, prefix) {
			return false
		}
	}
	return true
}
