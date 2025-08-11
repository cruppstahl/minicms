package cmd

import (
	"cms/core"
	"fmt"
)

func Version() {
	fmt.Printf("Version: %s\n", core.Version)
}
