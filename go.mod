module github.com/kuttiproject/kuttilib

go 1.16

require (
	github.com/kuttiproject/drivercore v0.2.0
	github.com/kuttiproject/kuttilog v0.1.2
	github.com/kuttiproject/workspace v0.2.2
)

retract [v0.1.0, v0.1.1] // Broke compatibility with original kutti
