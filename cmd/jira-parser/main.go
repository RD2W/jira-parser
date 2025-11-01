package main

import (
	"github.com/rd2w/jira-parser/internal/interfaces/cli"
	"github.com/rd2w/jira-parser/internal/version"
)

var (
	// BuildVersion can be set at build time using ldflags
	BuildVersion string
	// BuildCommit can be set at build time using ldflags
	BuildCommit string
	// BuildDate can be set at build time using ldflags
	BuildDate string
)

func init() {
	// If BuildVersion was set during build, use it
	if BuildVersion != "" {
		version.App.Version = BuildVersion
	}
	// If BuildCommit was set during build, use it
	if BuildCommit != "" {
		version.App.Commit = BuildCommit
	}
	// If BuildDate was set during build, use it
	if BuildDate != "" {
		version.App.Date = BuildDate
	}
}

func main() {
	cli.Execute()
}
