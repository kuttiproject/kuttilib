package kuttilib

import "github.com/kuttiproject/drivercore"

// Driver is a kutti driver
//
// Each Driver supports a number of Versions, and provides
// a driver-specific image for creating nodes. A Driver
// also provides a repository from which to fetch the list
// of supported versions and images.
type Driver struct {
	vmdriver drivercore.Driver
}

// Name returns the name of the driver.
func (d *Driver) Name() string {
	return d.vmdriver.Name()
}

// Description returns a one-line description of the driver.
func (d *Driver) Description() string {
	return d.vmdriver.Description()
}

// UsesNATNetworking returns true if the driver's networks use NAT,
// and therefore require node ports to be forwarded.
func (d *Driver) UsesNATNetworking() bool {
	return d.vmdriver.UsesNATNetworking()
}

// Status returns the driver status.
func (d *Driver) Status() string {
	return d.vmdriver.Status()
}

// UpdateVersionList fetches the latest list of available
// Versions for this driver, from the driver repository.
func (d *Driver) UpdateVersionList() error {
	return d.vmdriver.UpdateImageList()
}

// VersionNames returns the Kubernetes version strings
// of all available Versions for this driver.
func (d *Driver) VersionNames() []string {
	return d.vmdriver.K8sVersions()
}

// Versions returns the available Versions for this driver.
func (d *Driver) Versions() []*Version {
	rawimages, err := d.vmdriver.ListImages()
	if err != nil {
		return []*Version{}
	}

	result := make([]*Version, len(rawimages))

	for i := 0; i < len(rawimages); i++ {
		result[i] = &Version{
			image: rawimages[i],
		}
	}

	return result
}

// ForEachVersion iterates over available versions for this driver.
func (d *Driver) ForEachVersion(f func(*Version) bool) error {
	driver := d.vmdriver

	images, err := driver.ListImages()
	if err != nil {
		return err
	}

	for _, value := range images {
		version := &Version{image: value}
		cancel := f(version)
		if cancel {
			break
		}
	}

	return nil
}

// GetVersion gets the image for the specified Kubernetes version,
// or nil and an error.
func (d *Driver) GetVersion(version string) (*Version, error) {
	driver := d.vmdriver

	img, err := driver.GetImage(version)
	if err == nil {
		return &Version{image: img}, nil
	}

	return nil, err
}
