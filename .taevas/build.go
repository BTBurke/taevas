//go:build mage

package main

import (
	"fmt"
	"log"
)

// Use layout to register a template that serves as a base for other templates.
// See docs.
func Layout(template string) {

	if template == "" {
		log.Fatal("must specify template")
	}
	fmt.Printf("works for %s\n", template)
}
