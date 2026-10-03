package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	output := flag.String("output", "", "output swagger file")
	flag.Parse()
	if *output == "" || flag.NArg() == 0 {
		panic("output and at least one input are required")
	}
	var merged map[string]any
	paths := make(map[string]any)
	for _, filename := range flag.Args() {
		data, err := os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		var document map[string]any
		if err := json.Unmarshal(data, &document); err != nil {
			panic(err)
		}
		if merged == nil {
			merged = document
		}
		for path, operations := range document["paths"].(map[string]any) {
			if existing, exists := paths[path]; exists {
				if path == "/health" || strings.HasPrefix(path, "/health/") {
					continue
				}
				for method, operation := range operations.(map[string]any) {
					if _, duplicate := existing.(map[string]any)[method]; duplicate {
						panic(fmt.Sprintf("duplicate OpenAPI operation %s %s", method, path))
					}
					existing.(map[string]any)[method] = operation
				}
				continue
			}
			paths[path] = operations
		}
	}
	merged["paths"] = paths
	merged["schemes"] = []string{"http", "https"}
	merged["info"] = map[string]any{"title": "microservice-kit API", "version": "1.0", "description": "go-zero implementation of the native HTTP contract"}
	merged["securityDefinitions"] = map[string]any{"Bearer": map[string]any{"type": "apiKey", "name": "Authorization", "in": "header"}}
	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		panic(err)
	}
}
