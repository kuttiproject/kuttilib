package kuttilib

import "github.com/kuttiproject/kuttilog"

// VerbosityLevel represents the level of detail in logs.
type VerbosityLevel int

// Possible VerbosityLevel values are:
const (
	// VerbosityQuiet logs only minimal, machine-readable results
	VerbosityQuiet = VerbosityLevel(kuttilog.Quiet)
	// VerbosityInfo logs messages and human-readable results
	VerbosityInfo = VerbosityLevel(kuttilog.Info)
	// VerbosityDebug logs detailed debug information
	VerbosityDebug = VerbosityLevel(kuttilog.Debug)
)

// SetVerbosityLevel sets the current log level. If level is invalid, it is not changed.
func SetVerbosityLevel(level VerbosityLevel) {
	kuttilog.Setloglevel(int(level))
}
