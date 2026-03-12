// internal/openapi/schemas.go
package openapi

// my Api responses wrap the data into success, message and data so this reflects the sahpe of the api response on successful request
func buildEnvelopeSchema(dataSchemaName string) Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Schema{
			"success": {Type: "boolean"},
			"message": {Type: "string"},
			"data":    {Ref: "#/components/schemas/" + dataSchemaName},
		},
	}
}
