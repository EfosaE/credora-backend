// internal/openapi/metadata.go
package openapi

// Now in buildOperation, I need a way to say "the 200 response for GET /users/info returns a User", what it does and any info that may be important. Do this by adding a route-to-schema map:
// maps a "method:path" to the schema name it returns
type RouteMetadata struct {
	Summary     string
	Description string
	SchemaName  string // response schema, empty if none
}

var routeMetadata = map[string]RouteMetadata{
	"post:/auth/register": {
		Summary:     "Register a new user",
		Description: "Creates a new user account and returns a JWT token",
		SchemaName:  "AuthResponse",
	},
	"post:/auth/login": {
		Summary:     "Login",
		Description: "Authenticates a user and returns a JWT token",
		SchemaName:  "AuthResponse",
	},
	"get:/users/info": {
		Summary:     "Get current user",
		Description: "Returns the authenticated user's profile",
		SchemaName:  "User",
	},
	"get:/users/{email}": {
		Summary:     "Get the user detials via email",
		Description: "Returns the authenticated user's profile",
		SchemaName:  "User",
	},
	"get:/users/transactions": {
		Summary:     "Get transaction history",
		Description: "Returns paginated transaction history for the authenticated user",
		SchemaName:  "Transaction",
	},
	// add more as you build them out...
}
