package main

import (
	"fmt"

	"github.com/toon-format/toon-go"
)

type InventoryRow struct {
	SKU   string  `toon:"sku"`
	Qty   int     `toon:"qty"`
	Price float64 `toon:"price"`
}

type Inventory struct {
	Items []InventoryRow `toon:"items"`
}

func main() {
	payload := Inventory{
		Items: []InventoryRow{
			{SKU: "A-1", Qty: 5, Price: 9.99},
			{SKU: "B-2", Qty: 2, Price: 14.5},
		},
	}

	encoded, err := toon.MarshalString(
		payload,
		toon.WithDocumentDelimiter(toon.DelimiterPipe),
		toon.WithArrayDelimiter(toon.DelimiterPipe),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Pipe-delimited TOON:")
	fmt.Println(encoded)

	var decoded Inventory
	if err := toon.UnmarshalString(encoded, &decoded); err != nil {
		panic(err)
	}
	fmt.Printf("decoded %d rows\n", len(decoded.Items))
}
