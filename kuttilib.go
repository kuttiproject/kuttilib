package kuttilib

import (
	"regexp"
)

// ValidName checks for the validity of a name.
// Valid names of kutti objects are up to 10 characters long,
// must start with a lowercase letter, and may contain
// lowercase letters and digits only.
func ValidName(name string) bool {
	matched, _ := regexp.MatchString("^[a-z]([a-z0-9]{1,9})$", name)
	return matched
}

// ValidPort checks for the validity of a port number.
func ValidPort(portnumber int) bool {
	return portnumber > 0 && portnumber < 65536
}

// ValidateClusterName checks for the validity of a cluster name.
// It uses Validname to check name validity, and also checks if
// a cluster with that name already exists.
func ValidateClusterName(name string) error {
	if !ValidName(name) {
		return errInvalidName
	}

	// Check if name exists
	_, ok := config.Clusters[name]
	if ok {
		return errClusterExists
	}

	return nil
}
