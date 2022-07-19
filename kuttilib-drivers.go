package kuttilib

import (
	"github.com/kuttiproject/drivercore"
)

// ValidDriverName checks if the specified name belongs to
// an available Driver.
func ValidDriverName(drivername string) bool {
	return drivercore.IsRegisteredDriver(drivername)
}

// DriverNames returns the names of all available Drivers.
func DriverNames() []string {
	return drivercore.RegisteredDrivers()
}

// Drivers returns all available Drivers.
func Drivers() []*Driver {
	drivercount := drivercore.DriverCount()
	result := make([]*Driver, 0, drivercount)
	drivercore.ForEachDriver(func(vd drivercore.Driver) bool {
		result = append(result, &Driver{vmdriver: vd})
		return true
	})

	return result
}

// ForEachDriver iterates over Drivers.
//
// On each iteration, the callback function f is
// invoked with the Driver as a parameter. If the
// function returns false, iteration stops.
func ForEachDriver(f func(*Driver) bool) {
	drivercore.ForEachDriver(func(vd drivercore.Driver) bool {
		d := &Driver{vmdriver: vd}
		return f(d)
	})
}

// GetDriver gets the Driver with the specified name,
// or nil.
func GetDriver(drivername string) (*Driver, bool) {
	vd, ok := drivercore.GetDriver(drivername)
	if ok {
		return &Driver{vmdriver: vd}, ok
	}
	return nil, ok
}
