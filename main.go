package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BTBurke/taevas/build"
	"github.com/magefile/mage/mage"
)

func main() {

	root := build.GoRoot()

	if err := os.Chdir(filepath.Join(root, ".taevas")); err != nil {
		log.Fatal(err)
	}

	os.Exit(mage.Main())
}
