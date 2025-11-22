package utils

import (
	"encoding/json"
	"fmt"
)

// PrintJSON prints any struct or map as indented JSON
func PrintJSON(v any) {
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}
	fmt.Println(string(pretty))
}

// ReturnJSON Returns any struct or map as indented JSON
func ReturnJSON(v any) string {
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return ""
	}
	return string(pretty)
}
