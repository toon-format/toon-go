package main

import (
	"encoding/json"
	"fmt"

	"github.com/toon-format/toon-go"
)

type ToolResult struct {
	ToolName   string    `toon:"tool_name"`
	Timestamp  string    `toon:"timestamp"`
	RawPayload toon.TOON `toon:"raw_payload"`
}

func main() {
	fmt.Println("=== Tool Call Response Storage Example ===")
	fmt.Println()

	// Simulating a tool call that returns JSON
	fmt.Println("1. Tool returns JSON response:")
	toolResponse := `{"status":"success","data":{"count":42,"items":["a","b","c"]}}`
	fmt.Printf("JSON Response: %s\n\n", toolResponse)

	// Store the tool result with raw JSON payload
	result := ToolResult{
		ToolName:   "fetch_data",
		Timestamp:  "2025-11-17T16:30:00Z",
		RawPayload: toon.TOON(toolResponse), // Store JSON as-is
	}

	// Marshal the entire result to TOON
	fmt.Println("2. Store tool result as TOON:")
	encoded, err := toon.MarshalString(result)
	if err != nil {
		panic(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Later, unmarshal and access the raw payload
	fmt.Println("3. Retrieve and process:")
	var decoded ToolResult
	err = toon.UnmarshalString(encoded, &decoded)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tool Name: %s\n", decoded.ToolName)
	fmt.Printf("Timestamp: %s\n", decoded.Timestamp)
	fmt.Printf("Raw Payload (first 50 chars): %s...\n\n", decoded.RawPayload.String()[:50])

	// Process the JSON payload separately
	fmt.Println("4. Parse the JSON payload:")
	var jsonData map[string]any
	err = json.Unmarshal([]byte(decoded.RawPayload), &jsonData)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Status: %v\n", jsonData["status"])
	if data, ok := jsonData["data"].(map[string]any); ok {
		fmt.Printf("Count: %v\n", data["count"])
		fmt.Printf("Items: %v\n", data["items"])
	}

	fmt.Println()
	fmt.Println("=== Multiple Tool Calls ===")
	fmt.Println()

	// Simulating storing multiple tool call results
	type WorkflowLog struct {
		Name    string       `toon:"name"`
		Results []ToolResult `toon:"results"`
	}

	workflow := WorkflowLog{
		Name: "data_pipeline",
		Results: []ToolResult{
			{
				ToolName:   "fetch_data",
				Timestamp:  "2025-11-17T16:30:00Z",
				RawPayload: toon.TOON(`{"status":"success","count":42}`),
			},
			{
				ToolName:   "process_data",
				Timestamp:  "2025-11-17T16:31:00Z",
				RawPayload: toon.TOON(`{"status":"success","processed":42}`),
			},
			{
				ToolName:   "save_results",
				Timestamp:  "2025-11-17T16:32:00Z",
				RawPayload: toon.TOON(`{"status":"success","saved":true}`),
			},
		},
	}

	workflowEncoded, err := toon.MarshalString(workflow)
	if err != nil {
		panic(err)
	}
	fmt.Println("Workflow log as TOON:")
	fmt.Println(workflowEncoded)
}
