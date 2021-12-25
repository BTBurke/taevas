package build

var root string
var module string

// GoRoot returns the root of the current repository. Only works with go modules.
func GoRoot() string {
	// memoize root
	if root != "" {
		return root
	}
	r, err := Output("go", "list", "-f", "{{.Root}}")
	if err != nil {
		panic("taevas must be run in a go module")
	}
	root = r
	return root
}

// GoModule returns the name of the module
func GoModule() string {
	if module != "" {
		return module
	}
	m, err := Output("go", "list", "-f", "{{.Module}}")
	if err != nil {
		panic("taevas must be run in a go module")
	}
	module = m
	return module
}
