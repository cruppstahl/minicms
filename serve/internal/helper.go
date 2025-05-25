package internal

import (
	"encoding/json"
	"fmt"
	"log"
)

func PrettyPrint(context Context) {
	// Pretty-print the context using json.MarshalIndent
	contextJSON, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal context: %v", err)
	}
	fmt.Printf("Context:\n%s\n", string(contextJSON))
}
