module github.com/kuttiproject/kuttilib

go 1.24.2

require (
	github.com/kuttiproject/drivercore v0.3.1
	github.com/kuttiproject/kuttilog v0.2.1
	github.com/kuttiproject/workspace v0.3.1
)

retract [v0.1.0, v0.1.1] // Broke compatibility with original kutti
