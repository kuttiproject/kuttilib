package kuttilib

import (
	"sort"
	"time"

	"github.com/kuttiproject/kuttilog"

	"github.com/kuttiproject/drivercore"
)

// ClusterNames returns the names of all clusters. The
// order is not predictable.
func ClusterNames() []string {
	result := make([]string, len(config.Clusters))
	i := 0
	for clustername := range config.Clusters {
		result[i] = clustername
		i++
	}
	return result
}

// Clusters returns all clusters, sorted in reverse order of
// creation time.
func Clusters() []*Cluster {
	result := make([]*Cluster, len(config.Clusters))
	i := 0
	for _, cluster := range config.Clusters {
		result[i] = cluster
		i++
	}
	sort.Slice(result, func(a, b int) bool {
		return result[a].createdAt.After(result[b].createdAt)
	})
	return result
}

// ForEachCluster iterates over clusters.
func ForEachCluster(f func(*Cluster) bool) {
	for _, cluster := range config.Clusters {
		if cancel := f(cluster); cancel {
			break
		}
	}
}

// GetCluster gets a named cluster, or nil if not present.
func GetCluster(name string) (*Cluster, bool) {
	cluster, ok := config.Clusters[name]
	if !ok {
		return nil, ok
	}
	return cluster, ok
}

// DeleteCluster deletes a cluster.
// Currently, the cluster must be empty.
func DeleteCluster(clustername string, force bool) error {
	cluster, ok := GetCluster(clustername)
	if !ok {
		return errClusterDoesNotExist
	}

	if len(cluster.nodes) > 0 {
		return errClusterNotEmpty
	}

	if cluster.Driver().UsesPerClusterNetworking() {
		kuttilog.Println(kuttilog.Info, "Deleting network...")
		err := cluster.deletenetwork()
		if err != nil {
			if !force {
				return err
			}

			kuttilog.Printf(
				kuttilog.Quiet,
				"Warning: Errors returned while deleting network: %v. Some artifacts may need manual cleanup.",
				err,
			)
		}

		kuttilog.Println(kuttilog.Info, "Network deleted.")
	}

	delete(config.Clusters, clustername)

	return clusterconfigmanager.Save()
}

func newunmanagedcluster(name string, k8sversion string, drivername string) (*Cluster, error) {
	newCluster := &Cluster{
		name:       name,
		k8sVersion: k8sversion,
		driverName: drivername,
		createdAt:  time.Now(),
		nodes:      map[string]*Node{},
		status:     "UnInitialized",
	}

	// Ensure presence of Driver
	err := newCluster.ensuredriver()
	if err != nil {
		return newCluster, err
	}

	// Create Network if required
	if newCluster.Driver().UsesPerClusterNetworking() {
		kuttilog.Println(kuttilog.Info, "Creating network...")
		err = newCluster.createnetwork()
		if err != nil {
			return newCluster, err
		}

		kuttilog.Println(kuttilog.Info, "Network created.")
	}

	newCluster.clustertype = "Unmanaged"
	newCluster.status = "Ready"

	return newCluster, nil
}

// NewEmptyCluster creates a new, empty cluster.
// It uses ValidName to check name validity, and also checks if a cluster with the
// name already exists.
func NewEmptyCluster(name string, k8sversion string, drivername string) error {
	// Validate name
	err := ValidateClusterName(name)
	if err != nil {
		return err
	}

	// Validate driver
	driver, ok := drivercore.GetDriver(drivername)
	if !ok {
		return errDriverDoesNotExist
	}

	// Validate k8sversion
	driverimage, err := driver.GetImage(k8sversion)
	if err != nil {
		return err
	}

	if driverimage.Status() != drivercore.ImageStatusDownloaded {
		return errImageNotAvailable
	}

	if driverimage.Deprecated() {
		return errVersionDeprecated
	}

	// Create cluster
	newCluster, err := newunmanagedcluster(name, k8sversion, drivername)
	if err != nil {
		return err
	}

	config.Clusters[name] = newCluster
	return clusterconfigmanager.Save()
}
