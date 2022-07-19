package kuttilib

import (
	"encoding/json"

	"github.com/kuttiproject/drivercore"
)

// VersionStatus represents the current status of a version.
type VersionStatus drivercore.ImageStatus

// The VersionStatus* constants list valid version statuses.
const (
	VersionStatusNotDownloaded = VersionStatus(drivercore.ImageStatusNotDownloaded)
	VersionStatusDownloaded    = VersionStatus(drivercore.ImageStatusDownloaded)
)

// versiondata is a data-only representation of the Version type,
// used for serialization and output.
type versiondata struct {
	K8sVersion string
	Status     string
}

// Version represents a Kubernetes version that may be
// used to create a cluster.
type Version struct {
	image drivercore.Image
}

// K8sVersion returns the Kubernetes version string.
func (v *Version) K8sVersion() string {
	return v.image.K8sVersion()
}

// Status returns the local availability of the version.
func (v *Version) Status() VersionStatus {
	return VersionStatus(v.image.Status())
}

// MarshalJSON returns the JSON encoding of the version.
func (v *Version) MarshalJSON() ([]byte, error) {
	savedata := versiondata{
		K8sVersion: v.K8sVersion(),
		Status:     string(v.Status()),
	}

	return json.Marshal(savedata)
}

// Fetch downloads this version's image from the Driver
// repository.
func (v *Version) Fetch() error {
	return v.image.Fetch()
}

// FetchWithProgress downloads this version's image from the Driver
// repository into the local cache, and reports progress via the
// supplied callback. The callback reports current and total in bytes.
func (v *Version) FetchWithProgress(progress func(current int64, total int64)) error {
	return v.image.FetchWithProgress(progress)
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
