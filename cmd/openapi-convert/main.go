// Command openapi-convert converts OpenAPI 3.1 specs to 3.0.3 for ogen compatibility.
//
// Usage:
//
//	go run ./cmd/openapi-convert openapi/openapi-v3.1.json openapi/openapi-v3.0.json
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input-v3.1.json> <output-v3.0.json>\n", os.Args[0]) //nolint:gosec // G705: CLI tool writing to stderr, not browser context
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Read raw JSON
	data, err := os.ReadFile(inputFile) //nolint:gosec // G703: CLI tool, user-specified path is expected
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse as generic map
	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	originalVersion := spec["openapi"]
	fmt.Printf("Input OpenAPI version: %v\n", originalVersion)

	// Convert to 3.0.3
	spec["openapi"] = "3.0.3"

	// Remove 3.1-only top-level fields
	delete(spec, "webhooks")

	// Fix all schemas recursively
	fixValue(spec)

	// Marshal back to JSON
	output, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(outputFile, output, 0600); err != nil { //nolint:gosec // G703: CLI tool, user-specified path is expected
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Output OpenAPI version: 3.0.3\n")
	fmt.Printf("Wrote converted spec to: %s\n", outputFile)
}

// fixValue recursively processes JSON values to convert 3.1 -> 3.0 patterns
func fixValue(v any) {
	switch val := v.(type) {
	case map[string]any:
		fixObject(val)
	case []any:
		for _, item := range val {
			fixValue(item)
		}
	}
}

// fixObject processes a JSON object to convert 3.1 -> 3.0 patterns
func fixObject(obj map[string]any) {
	// Handle exclusiveMinimum (number in 3.1 -> remove, use minimum)
	if exMin, ok := obj["exclusiveMinimum"]; ok {
		if _, isNum := exMin.(float64); isNum {
			// In 3.1, exclusiveMinimum is a number (the exclusive bound)
			// In 3.0, exclusiveMinimum is a boolean
			// Convert: remove exclusiveMinimum, keep minimum if exists
			delete(obj, "exclusiveMinimum")
		}
	}

	// Handle exclusiveMaximum (number in 3.1 -> remove, use maximum)
	if exMax, ok := obj["exclusiveMaximum"]; ok {
		if _, isNum := exMax.(float64); isNum {
			delete(obj, "exclusiveMaximum")
		}
	}

	// Handle const (3.1) -> enum with single value (3.0)
	if constVal, ok := obj["const"]; ok {
		obj["enum"] = []any{constVal}
		delete(obj, "const")
	}

	// Handle anyOf with type:null (convert to nullable:true)
	if anyOf, ok := obj["anyOf"].([]any); ok {
		var nonNullSchemas []any
		hasNull := false

		for _, schema := range anyOf {
			if schemaMap, ok := schema.(map[string]any); ok {
				if schemaMap["type"] == "null" {
					hasNull = true
				} else {
					nonNullSchemas = append(nonNullSchemas, schema)
				}
			} else {
				nonNullSchemas = append(nonNullSchemas, schema)
			}
		}

		if hasNull {
			if len(nonNullSchemas) == 1 {
				// Replace anyOf with the single non-null schema + nullable:true
				if schemaMap, ok := nonNullSchemas[0].(map[string]any); ok {
					delete(obj, "anyOf")
					for k, v := range schemaMap {
						obj[k] = v
					}
					obj["nullable"] = true
				}
			} else if len(nonNullSchemas) > 1 {
				// Keep anyOf but remove null and add nullable
				obj["anyOf"] = nonNullSchemas
				obj["nullable"] = true
			} else {
				// Only null type - convert to nullable any
				delete(obj, "anyOf")
				obj["nullable"] = true
			}
		}
	}

	// Handle oneOf with type:null similarly
	if oneOf, ok := obj["oneOf"].([]any); ok {
		var nonNullSchemas []any
		hasNull := false

		for _, schema := range oneOf {
			if schemaMap, ok := schema.(map[string]any); ok {
				if schemaMap["type"] == "null" {
					hasNull = true
				} else {
					nonNullSchemas = append(nonNullSchemas, schema)
				}
			} else {
				nonNullSchemas = append(nonNullSchemas, schema)
			}
		}

		if hasNull {
			if len(nonNullSchemas) == 1 {
				if schemaMap, ok := nonNullSchemas[0].(map[string]any); ok {
					delete(obj, "oneOf")
					for k, v := range schemaMap {
						obj[k] = v
					}
					obj["nullable"] = true
				}
			} else if len(nonNullSchemas) > 1 {
				obj["oneOf"] = nonNullSchemas
				obj["nullable"] = true
			} else {
				delete(obj, "oneOf")
				obj["nullable"] = true
			}
		}
	}

	// Handle type as array (3.1 feature): ["string", "null"] -> type: "string", nullable: true
	if typeVal, ok := obj["type"].([]any); ok {
		var nonNullTypes []string
		hasNull := false

		for _, t := range typeVal {
			if ts, ok := t.(string); ok {
				if ts == "null" {
					hasNull = true
				} else {
					nonNullTypes = append(nonNullTypes, ts)
				}
			}
		}

		if len(nonNullTypes) == 1 {
			obj["type"] = nonNullTypes[0]
			if hasNull {
				obj["nullable"] = true
			}
		} else if len(nonNullTypes) > 1 {
			// Multiple types - this is complex, use anyOf
			var schemas []any
			for _, t := range nonNullTypes {
				schemas = append(schemas, map[string]any{"type": t})
			}
			delete(obj, "type")
			obj["anyOf"] = schemas
			if hasNull {
				obj["nullable"] = true
			}
		}
	}

	// Recurse into all values
	for _, v := range obj {
		fixValue(v)
	}
}
