package version

// Info contains version information
type Info struct {
	Version string
	Commit  string
	Date    string
}

// App contains the version information for the application
var App = Info{
	Version: "v1.0.0",
}
