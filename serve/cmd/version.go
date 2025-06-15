package cmd

import (
	"fmt"
	"serve/core"
)

func Version() {
	fmt.Printf("Version: %s\n", core.Version)
}
