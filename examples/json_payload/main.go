package main

import (
	"encoding/json"
	"fmt"

	"github.com/toon-format/toon-go"
)

type Response struct {
	EventID string    `json:"event_id"`
	Payload toon.TOON `json:"payload"`
}

func main() {
	fmt.Println("=== JSON Payload Unmarshaling Example ===")
	fmt.Println()

	// Tool call returns JSON with nested payload
	jsonResponse := `{
  "event_id": "evt_xyz",
  "payload": {
    "id": "order_xyz",
    "amount": 99.99,
    "items": ["widget", "gadget"]
  }
}`

	fmt.Println("1. Incoming JSON from tool call:")
	fmt.Println(jsonResponse)
	fmt.Println()

	// Unmarshal JSON - payload field automatically stored as TOON
	fmt.Println("2. Unmarshal JSON into Go struct with toon.TOON field:")
	var response Response
	err := json.Unmarshal([]byte(jsonResponse), &response)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Event ID: %s\n", response.EventID)
	fmt.Printf("Payload (raw): %s\n", response.Payload.String())
	fmt.Println()

	// Process the payload whenever needed - it's stored as TOON
	fmt.Println("3. Process the payload (decode from TOON):")
	payloadData, err := toon.Decode(response.Payload)
	if err != nil {
		panic(err)
	}

	// Access the decoded data
	payloadMap := payloadData.(map[string]any)
	fmt.Printf("Order ID: %s\n", payloadMap["id"])
	fmt.Printf("Amount: %.2f\n", payloadMap["amount"])
	fmt.Printf("Items: %v\n", payloadMap["items"])
	fmt.Println()

	// Example: Store payload in database
	fmt.Println("4. Store TOON payload in database:")
	fmt.Printf("TOON format (length %d): %s\n", len(response.Payload), response.Payload.String())
	fmt.Println()

	// Example: Round-trip through JSON
	fmt.Println("5. Round-trip: struct → JSON → struct:")
	jsonBytes, _ := json.Marshal(response)
	fmt.Printf("JSON output: %s\n", string(jsonBytes))

	var roundtrip Response
	json.Unmarshal(jsonBytes, &roundtrip)
	fmt.Printf("Payload after round-trip: %s\n", roundtrip.Payload.String())
	fmt.Println()

	// Example: Multiple tool responses
	fmt.Println("6. Multiple tool call responses:")
	multipleJSON := `[
		{"event_id": "evt_001", "payload": {"type":"order","status":"pending","total":250.00}},
		{"event_id": "evt_002", "payload": {"type":"payment","method":"card","amount":150.00}},
		{"event_id": "evt_003", "payload": {"type":"shipment","tracking":"TRACK123","carrier":"UPS"}}
	]`

	var responses []Response
	json.Unmarshal([]byte(multipleJSON), &responses)

	for _, resp := range responses {
		decoded, _ := toon.Decode(resp.Payload)
		payload := decoded.(map[string]any)
		fmt.Printf("Event %s: type=%s\n", resp.EventID, payload["type"])
	}
}
