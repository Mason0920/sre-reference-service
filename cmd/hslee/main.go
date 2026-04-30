package main

import (
	"fmt"
	"os"
)

// Version metadata injected at build time via -ldflags.
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	fmt.Fprintf(os.Stdout, "hslee SRE foundation — version=%s commit=%s built=%s\n", version, commit, buildTime)
}
