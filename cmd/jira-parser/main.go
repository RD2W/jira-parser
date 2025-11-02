package main

import (
	"github.com/rd2w/jira-parser/internal/interfaces/cli"
	"github.com/rd2w/jira-parser/internal/version"
)

var (
	// BuildVersion can be set at build time using ldflags
	buildVersion string
	// BuildCommit can be set at build time using ldflags
	buildCommit string
	// BuildDate can be set at build time using ldflags
	buildDate string
)

func init() {
	// If BuildVersion was set during build, use it
	if buildVersion != "" {
		version.App.Version = buildVersion
	}
	// If BuildCommit was set during build, use it
	if buildCommit != "" {
		version.App.Commit = buildCommit
	}
	// If BuildDate was set during build, use it
	if buildDate != "" {
		version.App.Date = buildDate
	}
}

func main() {
	cli.Execute()
}
