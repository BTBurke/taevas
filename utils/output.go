package utils

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
)

// Output runs a command and returns the output from stdout.  The environment of the current
// shell is copied to the subshell.
func Output(cmd string, args ...string) (string, error) {
	// fix for mage/sh which logs commands by default
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	env := make(map[string]string)
	for _, s := range os.Environ() {
		kv := strings.Split(s, "=")
		env[kv[0]] = kv[1]
	}
	return sh.OutputWith(env, cmd, args...)
}
