module github.com/kuttiproject/kuttilib

go 1.18

require (
	github.com/kuttiproject/drivercore v0.3.0
	github.com/kuttiproject/kuttilog v0.2.0
	github.com/kuttiproject/workspace v0.3.0
)

retract [v0.1.0, v0.1.1] // Broke compatibility with original kutti
