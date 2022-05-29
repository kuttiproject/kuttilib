package kuttilib

import (
	"encoding/json"
	"sort"

	"github.com/kuttiproject/drivercore"
)

// driverdata is a data-only representation of the Driver type,
// used for serialization and output.
type driverdata struct {
	Name                     string
	Description              string
	UsesPerClusterNetworking bool
	UsesNATNetworking        bool
	Status                   string
	Error                    string
}

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

// UsesPerClusterNetworking returns true if the driver creates virtual networks
// per cluster.
func (d *Driver) UsesPerClusterNetworking() bool {
	return d.vmdriver.UsesPerClusterNetworking()
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

// Error returns the last error in the driver.
func (d *Driver) Error() string {
	return d.vmdriver.Error()
}

// MarshalJSON returns the JSON encoding of the driver.
func (d *Driver) MarshalJSON() ([]byte, error) {
	savedata := driverdata{
		Name:                     d.Name(),
		Description:              d.Description(),
		UsesPerClusterNetworking: d.UsesPerClusterNetworking(),
		UsesNATNetworking:        d.UsesNATNetworking(),
		Status:                   d.Status(),
		Error:                    d.Error(),
	}

	return json.Marshal(savedata)
}

// UpdateVersionList fetches the latest list of available
// Versions for this driver, from the driver repository.
func (d *Driver) UpdateVersionList() error {
	return d.vmdriver.UpdateImageList()
}

// VersionNames returns the Kubernetes version strings
// of all available Versions for this driver, in
// ascending order of K8sVersion.
func (d *Driver) VersionNames() []string {
	result := d.vmdriver.K8sVersions()
	sort.Strings(result)
	return result
}

// Versions returns the available Versions for this driver,
// in ascending order of K8sVersion.
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

	sort.Slice(result, func(i, j int) bool {
		return result[i].K8sVersion() < result[j].K8sVersion()
	})
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
