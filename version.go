package kuttilib

import "github.com/kuttiproject/drivercore"

// VersionStatus represents the current status of a version image.
type VersionStatus drivercore.ImageStatus

// Possible VersionStatus values are:
const (
	VersionStatusNotDownloaded = VersionStatus(drivercore.ImageStatusNotDownloaded)
	VersionStatusDownloaded    = VersionStatus(drivercore.ImageStatusDownloaded)
)

// Version represents a Kubernetes version that may be
// used to create a cluster.
type Version struct {
	image drivercore.Image
}

// K8sversion returns the Kubernetes version string.
func (v *Version) K8sversion() string {
	return v.image.K8sVersion()
}

// Status returns the local availability of the version.
func (v *Version) Status() VersionStatus {
	return VersionStatus(v.image.Status())
}

// Fetch downloads this version's image from the Driver
// repository.
func (v *Version) Fetch() error {
	return v.image.Fetch()
}

// FromFile imports this version's image from the specified
// local file.
func (v *Version) FromFile(filename string) error {
	return v.image.FromFile(filename)
}

// PurgeLocal removes the local cached copy of a version.
func (v *Version) PurgeLocal() error {
	return v.image.PurgeLocal()
}
