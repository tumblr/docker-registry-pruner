package main

import (
	"fmt"
	"io"
	"runtime"

	"github.com/tumblr/docker-registry-pruner/internal/pkg/version"
)

// ShowVersion shows the version of the jawner
func ShowVersion(o io.Writer) {
	fmt.Fprintf(o, "docker-registry-pruner version:%s (commit:%s branch:%s) built on %s with %s\n", version.Version, version.Commit, version.Branch, version.BuildDate, runtime.Version())
}
