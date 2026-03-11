// internal/openapi/schemas.go
package openapi

// DefineSchemas returns all your model definitions.
// You add a new entry here for every model you want documented.
func DefineSchemas() map[string]Schema {
	return map[string]Schema{
		"User": {
			Type: "object",
			Properties: map[string]Schema{
				"id":        {Type: "string", Format: "uuid"},
				"email":     {Type: "string"},
				"name":      {Type: "string"},
				"createdAt": {Type: "string", Format: "date-time"},
			},
		},
		"Transaction": {
			Type: "object",
			Properties: map[string]Schema{
				"id":        {Type: "string", Format: "uuid"},
				"amount":    {Type: "number"},
				"status":    {Type: "string"},
				"createdAt": {Type: "string", Format: "date-time"},
			},
		},
		// add more models here as you build them out
	}
}
