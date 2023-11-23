package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// This is a minimal implementation of the external resource protocol
// intended only for use in the provider acceptance tests.
//
// In practice it's likely not much harder to just write a real Terraform
// plugin if you're going to be writing your resource in Go anyway;
// this example is just in Go because we want to avoid introducing
// additional language runtimes into the test environment.
func main() {
	queryBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	query := map[string]any{}
	err = json.Unmarshal(queryBytes, &query)
	if err != nil {
		panic(err)
	}

	if query["fail"] == "true" {
		fmt.Fprintf(os.Stderr, "I was asked to fail\n")
		os.Exit(1)
	}

	var result = map[string]any{
		"result":      "yes",
		"query_value": query["value"],
	}

	if len(os.Args) >= 2 {
		result["argument"] = os.Args[1]
	}

	for queryKey, queryValue := range query {
		if queryKey == "old_result" {
			continue
		}
		result[queryKey] = queryValue
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(resultBytes)
	os.Exit(0)
}
