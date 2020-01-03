package version

import (
	"fmt"
	"runtime"
)

var (
	// set via build args
	gitVersion = "v0.0.0-master"
	gitCommit  string
	buildDate  string
)

// Info is a container for version information.
type Info struct {
	GitVersion string
	GitCommit  string
	BuildDate  string
	GoVersion  string
	Compiler   string
	Platform   string
}

// Get returns the version info.
func Get() *Info {
	return &Info{
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
