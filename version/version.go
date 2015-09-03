package version

var (
	Version = "0.1.0"
	// GitCommit is updated upon deploy
	GitCommit = "HEAD"
)

func FullVersion() string {
	return Version + " (" + GitCommit + ")"
}
