// Command configdoc prints the generated configuration reference.
package main

import (
	"fmt"
	"os"

	"github.com/aleogr/iam/shared/config"
)

func main() {
	if err := config.Document(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "configdoc:", err)
		os.Exit(1)
	}
}
