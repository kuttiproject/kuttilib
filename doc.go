// Package kuttilib provides an API to manage small, non-production
// Kubernetes clusters.
//
// This is an abstract API, expected to work on top of multiple
// platforms such as local hypervisors or cloud providers. The
// abstraction is maintained through "drivers", which implement the
// actual functionality of setting up networks and hosts on which
// to run Kubernetes.
//
// A kutti client can be created by referencing this package together
// with one or more drivers.
//
// Workspaces
//
// The kuttilib package stores cluster and node configuration in
// configuration files contained in specific directories within a
// workspace. Drivers may also use directories within a workspace
// to store images which provide a template to create nodes running
// a specific Kubernetes version. So, it is important to select the
// correct workspace before using the API in any way.
//
// The default workspace is assumed to be the home directory of the
// operating system user who invokes the API. This can be changed
// using the SetWorkspace method. See that method, and the package
// github.com/kuttiproject/workspace for details.
//
// Drivers
//
// The kuttilib package itself does not include any drivers. When
// a kutti client package references any driver packages, these
// drivers will automatically become available to the kuttilib API.
// See the Driver* family of functions and the Driver type for
// details.
//
// Versions
//
// Each driver is responsible for providing template images for one
// or more Kubernetes versions. A driver's UpdateVersionList method
// may be called to get the latest list from a driver-provided
// repository. The Fetch and FromFile methods provide ways to
// import these images into a workspace, so that clusters can be
// created. See the Driver type for details.
//
// Clusters
//
// One or more clusters can be created and maintained in a workspace
// once the appropriate version templates have been downloaded. See
// the Cluster family of functions and the Cluster type for details.
//
// Nodes
//
// Nodes may be created and managed for each cluster. See the Cluster
// and Node types for details.
package kuttilib
