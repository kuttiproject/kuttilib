package kuttilib

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/kuttiproject/kuttilog"

	"github.com/kuttiproject/drivercore"
)

// clusterdata is a data-only representation of the Cluster type,
// used for serialization and output.
type clusterdata struct {
	Name       string
	DriverName string
	K8sVersion string
	CreatedAt  time.Time
	Type       string
	Nodes      map[string]*Node
}

// Cluster represents a Kubernetes cluster, consisting of Nodes.
//
// Each Cluster is associated with a Driver, which will handle
// the actual creation of networks and hosts on the underlying
// platform.
//
// Each cluster is also associated with a Version, which
// represents the version of Kubernetes this cluster will run.
//
// These decisions are  taken while creating the cluster, and
// cannot be changed later.
type Cluster struct {
	name        string
	driverName  string
	driver      drivercore.Driver
	k8sVersion  string
	createdAt   time.Time
	network     drivercore.Network
	nodes       map[string]*Node
	clustertype string
	status      string
}

// Name returns the name of the cluster.
func (c *Cluster) Name() string {
	return c.name
}

// DriverName returns the name of the Driver
// associated with this cluster.
func (c *Cluster) DriverName() string {
	return c.driverName
}

// Driver returns the Driver associated with
// this cluster.
func (c *Cluster) Driver() *Driver {
	result, _ := GetDriver(c.driverName)
	return result
}

// K8sVersion returns the Kubernetes version
// associated with this cluster.
func (c *Cluster) K8sVersion() string {
	return c.k8sVersion
}

// CreatedAt returns the time this cluster was created.
func (c *Cluster) CreatedAt() time.Time {
	return c.createdAt
}

// Type returns the type of this cluster.
func (c *Cluster) Type() string {
	return c.clustertype
}

// ValidateNodeName checks for the validity of a node name.
// It uses Validname to check name validity, and also checks if a node name
// already exists in the cluster.
func (c *Cluster) ValidateNodeName(name string) error {
	if !ValidName(name) {
		return errInvalidName
	}

	// Check if name exists
	_, ok := c.nodes[name]
	if ok {
		return errNodeExists
	}

	return nil
}

// NodeNames returns the names of all nodes
// in the cluster. The order is not predictable.
func (c *Cluster) NodeNames() []string {
	result := make([]string, len(c.nodes))
	i := 0
	for nodename := range c.nodes {
		result[i] = nodename
		i++
	}
	return result
}

// Nodes returns all the Nodes in the cluster, in reverse
// order of creation time.
func (c *Cluster) Nodes() []*Node {
	result := make([]*Node, len(c.nodes))
	i := 0
	for _, node := range c.nodes {
		result[i] = node
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].createdAt.After(result[j].createdAt)
	})
	return result
}

// ForEachNode iterates over Nodes in the cluster.
//
// On each iteration, the callback function f is
// invoked with the Node as a parameter. If the
// function returns false, iteration stops.
func (c *Cluster) ForEachNode(f func(*Node) bool) {
	for _, node := range c.nodes {
		if cancel := f(node); cancel {
			break
		}
	}
}

// GetNode returns the node with the specified name, or nil.
func (c *Cluster) GetNode(nodename string) (*Node, bool) {
	result, ok := c.nodes[nodename]
	return result, ok
}

// DeleteNode deletes a node completely. By default, a node is not deleted
// if it is running. The force parameter causes the node to be stopped and
// deleted. In some rare cases for some drivers, manual cleanup may be
// needed after a forced delete.
func (c *Cluster) DeleteNode(nodename string, force bool) error {
	n, ok := c.nodes[nodename]
	if !ok {
		return errNodeNotFound
	}

	nodestatus := n.Status()

	if nodestatus == NodeStatusUnknown ||
		nodestatus == NodeStatusError {

		return c.deletenodeentry(nodename)
	}

	if nodestatus == NodeStatusRunning {
		if !force {
			return errNodeIsRunning
		}

		kuttilog.Printf(kuttilog.Info, "Stopping node %s...", nodename)
		err := n.ForceStop()
		if err != nil {
			kuttilog.Printf(kuttilog.Quiet, "Error stopping node: %v. Some VirtualBox artifacts may be left behind.")
		} else {
			kuttilog.Printf(kuttilog.Info, "Node %s stopped.", nodename)
		}
	}

	// Unmap ports
	kuttilog.Println(kuttilog.Info, "Unmapping ports...")
	for key := range n.ports {
		err := n.host.UnforwardPort(key)
		if err != nil {
			kuttilog.Printf(kuttilog.Quiet, "Error while unmapping ports for node '%s': %v.", nodename, err)
		}
	}
	kuttilog.Println(kuttilog.Info, "Ports unmapped.")

	return c.deletenode(nodename, force)
}

// NewUninitializedNode adds a node, but does not join it to a kubernetes cluster.
// It uses Validname to check name validity, and also checks if a node with the
// name already exists.
func (c *Cluster) NewUninitializedNode(nodename string) (*Node, error) {
	err := c.ValidateNodeName(nodename)
	if err != nil {
		return nil, err
	}

	return c.addnode(nodename, "Unmanaged")
}

// CheckHostport returns an error if a host port is occupied in the current cluster.
func (c *Cluster) CheckHostport(hostport int) error {
	for _, nodevalue := range c.nodes {
		for _, hostportvalue := range nodevalue.ports {
			if hostportvalue == hostport {
				return errPortHostPortAlreadyUsed
			}
		}
	}
	return nil
}

// MarshalJSON returns the JSON encoding of the cluster.
func (c *Cluster) MarshalJSON() ([]byte, error) {
	utcloc, _ := time.LoadLocation("UTC")
	savedata := clusterdata{
		Name:       c.name,
		DriverName: c.driverName,
		K8sVersion: c.k8sVersion,
		CreatedAt:  c.createdAt.In(utcloc),
		Nodes:      c.nodes,
		Type:       c.clustertype,
	}

	return json.Marshal(savedata)
}

// UnmarshalJSON  parses and restores a JSON-encoded
// cluster.
func (c *Cluster) UnmarshalJSON(b []byte) error {
	var loaddata clusterdata

	err := json.Unmarshal(b, &loaddata)
	if err != nil {
		return err
	}

	localloc, _ := time.LoadLocation("Local")

	c.name = loaddata.Name
	c.driverName = loaddata.DriverName
	c.k8sVersion = loaddata.K8sVersion
	c.createdAt = loaddata.CreatedAt.In(localloc)
	c.nodes = loaddata.Nodes
	c.clustertype = loaddata.Type

	return nil
}

func (c *Cluster) ensuredriver() error {
	if c.driver == nil {
		driver, ok := drivercore.GetDriver(c.driverName)
		if !ok {
			c.status = "DriverNotPresent"
			return errDriverDoesNotExist
		}

		c.driver = driver
		c.status = "Driver" + c.driver.Status()
	}

	return nil
}

func (c *Cluster) createnetwork() error {
	nw, err := c.driver.NewNetwork(c.name)
	if err != nil {
		c.status = "NetworkError"
		return err
	}
	c.network = nw
	c.status = "NetworkReady"
	return nil
}

func (c *Cluster) deletenetwork() error {
	c.ensuredriver()
	err := c.driver.DeleteNetwork(c.name)
	if err != nil {
		c.status = "NetworkDeleteError"
		return err
	}
	c.network = nil
	c.status = "NetworkDeleted"
	return nil
}

func (c *Cluster) addnode(nodename string, nodetype string) (*Node, error) {
	err := c.ensuredriver()
	if err != nil {
		return nil, err
	}

	newnode := &Node{
		cluster:     c,
		clusterName: c.name,
		name:        nodename,
		createdAt:   time.Now(),
		nodetype:    nodetype,
		ports:       map[int]int{},
	}

	err = newnode.createhost()
	if err == nil {
		c.nodes[nodename] = newnode
		err = clusterconfigmanager.Save()
	}

	return newnode, err
}

func (c *Cluster) deletenodeentry(nodename string) error {
	delete(c.nodes, nodename)
	return clusterconfigmanager.Save()
}

func (c *Cluster) deletenode(nodename string, force bool) error {
	err := c.ensuredriver()
	if err != nil {
		return err
	}

	err = c.driver.DeleteMachine(nodename, c.name)
	if err == nil || force {
		err = c.deletenodeentry(nodename)
	}

	return err
}
