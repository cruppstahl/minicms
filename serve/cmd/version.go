package cmd

import (
	"fmt"
	"serve/core"
)

func RunVersion() {
	fmt.Printf("Version: %s\n", core.Version)
}
