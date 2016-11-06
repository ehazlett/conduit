package version

var (
	name        = "conduit"
	version     = "0.6.0"
	description = "docker container deployment"
	GitCommit   = "HEAD"
)

func Name() string {
	return name
}

func Version() string {
	return version + " (" + GitCommit + ")"
}

func Description() string {
	return description
}

func FullVersion() string {
	return Name() + " " + Version()
}
