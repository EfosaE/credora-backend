package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// WorkerID generates a unique worker consumer name.
//
// domain represents the service domain, e.g.:
// "email", "push", "notification", "transfer"
func WorkerID(domain string) string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown-host"
	}

	return fmt.Sprintf("%s-%s-%d",
		domain,
		hostname,
		os.Getpid(),
	)
}

//Returns typed struct to map types
func StructToMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return result, nil
}