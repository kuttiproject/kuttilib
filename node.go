package kuttilib

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kuttiproject/kuttilog"

	"github.com/kuttiproject/drivercore"
)

// NodeStatus represents the status of a Node.
type NodeStatus drivercore.MachineStatus

// The NodeStatus* constants define valid node statuses.
const (
	NodeStatusError   = NodeStatus(drivercore.MachineStatusError)
	NodeStatusUnknown = NodeStatus(drivercore.MachineStatusUnknown)
	NodeStatusStopped = NodeStatus(drivercore.MachineStatusStopped)
	NodeStatusRunning = NodeStatus(drivercore.MachineStatusRunning)
)

// nodedata is a data-only representation of the Node type,
// used for serialization and output.
type nodedata struct {
	ClusterName string
	Name        string
	CreatedAt   time.Time
	Type        string
	Ports       map[int]int
}

// Node represents a node in a Kubernetes cluster.
//
// The associated Cluster's Driver ensures that an appropriate
// host is created and manipulated per node.
type Node struct {
	cluster     *Cluster
	clusterName string
	name        string
	createdAt   time.Time
	nodetype    string
	host        drivercore.Machine
	//status      string
	ports map[int]int
}

// Name returns the name of the node.
func (n *Node) Name() string {
	return n.name
}

// Type returns the type of this node.
func (n *Node) Type() string {
	return n.nodetype
}

// CreatedAt returns the time this node was created.
func (n *Node) CreatedAt() time.Time {
	return n.createdAt
}

// Ports returns the node port to host port
// mappings of this node.
func (n *Node) Ports() map[int]int {
	return n.ports
}

// IPAddress returns the IP address of the node if it is running,
// or an empty string if not.
func (n *Node) IPAddress() string {
	err := n.ensurehost()
	if err != nil {
		return ""
	}

	return n.host.IPAddress()
}

// SSHAddress returns the address to SSH into the node.
// The return value is in "HOST:PORT" format if the node
// is running, or an empty string if not. If the Driver
// uses NAT networking, the host is "localhost".
func (n *Node) SSHAddress() string {
	if n.Cluster().Driver().UsesNATNetworking() {
		port, ok := n.Ports()[22]
		if !ok {
			return ""
		}
		return fmt.Sprintf("localhost:%v", port)

	}
	err := n.ensurehost()
	if err != nil {
		return ""
	}

	return n.host.SSHAddress()
}

// Cluster returns the Cluster this node belongs to.
func (n *Node) Cluster() *Cluster {
	if n.cluster == nil {
		n.cluster = config.Clusters[n.clusterName]
		n.cluster.ensuredriver()
	}
	return n.cluster
}

// Status returns the current status of this node.
func (n *Node) Status() NodeStatus {
	err := n.ensurehost()
	if err != nil {
		return NodeStatusError
	}

	return NodeStatus(n.host.Status())
}

// Start starts this node.
func (n *Node) Start() error {
	err := n.ensurehost()
	if err != nil {
		return err
	}

	if n.Status() == NodeStatusStopped {
		err = n.host.Start()
		if err != nil {
			return err
		}

		n.host.WaitForStateChange(25)
		return nil
	}

	return errNodeCannotStart
}

// ForceStart tries to forcibly start this node.
// It does not check the current status before doing so.
func (n *Node) ForceStart() error {
	err := n.ensurehost()
	if err != nil {
		return err
	}

	err = n.host.Start()
	if err != nil {
		return err
	}

	// TODO: Consider moving this wait, or standardize the duration
	kuttilog.Print(kuttilog.Info, "Waiting for node to start...")
	n.host.WaitForStateChange(25)
	kuttilog.Println(kuttilog.Info, "Done.")
	return nil
}

// Stop stops this node gracefully.
func (n *Node) Stop() error {
	err := n.ensurehost()
	if err != nil {
		return err
	}

	if n.Status() == NodeStatusRunning {
		err = n.host.Stop()
		if err != nil {
			return err
		}

		n.host.WaitForStateChange(25)
		return nil
	}

	return errNodeCannotStop

}

// ForceStop tries to forcibly stop this node.
// It does not check the current status before doing so.
func (n *Node) ForceStop() error {
	err := n.ensurehost()
	if err != nil {
		return err
	}

	err = n.host.ForceStop()
	if err != nil {
		return err
	}

	// TODO: Consider moving this wait, or standardize the duration
	kuttilog.Print(kuttilog.Info, "Waiting for node to stop...")
	n.host.WaitForStateChange(25)
	kuttilog.Println(kuttilog.Info, "Done.")
	return nil
}

// ForwardSSHPort forwards the node's SSH port to the specified host
// port.
func (n *Node) ForwardSSHPort(hostport int) error {
	return n.ForwardPort(hostport, 22)
}

// ForwardPort forwards a port of the node to the specified host port.
func (n *Node) ForwardPort(hostport int, nodeport int) error {
	err := n.Cluster().ensuredriver()
	if err != nil {
		return err
	}

	if !n.Cluster().driver.UsesNATNetworking() {
		return errPortForwardNotSupported
	}

	if !ValidPort(nodeport) {
		return errPortNodePortInvalid
	}

	if !ValidPort(hostport) {
		return errPortHostPortInvalid
	}

	err = n.ensurehost()
	if err != nil {
		return err
	}

	err = n.CheckHostPort(hostport)
	if err != nil {
		return err
	}

	_, ok := n.ports[nodeport]
	if ok {
		return errPortNodePortInUse
	}

	err = n.host.ForwardPort(hostport, nodeport)
	if err != nil {
		return err
	}

	n.ports[nodeport] = hostport
	return clusterconfigmanager.Save()
}

// UnforwardPort removes any mapping of the specified node port.
func (n *Node) UnforwardPort(nodeport int) error {
	cluster := n.Cluster()
	err := cluster.ensuredriver()
	if err != nil {
		return err
	}

	if !cluster.driver.UsesNATNetworking() {
		return errPortForwardNotSupported
	}

	if !ValidPort(nodeport) {
		return errPortNodePortInvalid
	}

	if nodeport == 22 {
		return errPortCannotUnmap
	}

	_, ok := n.ports[nodeport]
	if !ok {
		return errPortNotForwarded
	}

	err = n.ensurehost()
	if err != nil {
		return err
	}

	err = n.host.UnforwardPort(nodeport)
	if err != nil {
		return err
	}

	delete(n.ports, nodeport)
	return clusterconfigmanager.Save()
}

// CheckHostPort checks if a host port is occupied in the current cluster.
func (n *Node) CheckHostPort(hostport int) error {
	c := n.Cluster()
	return c.CheckHostPort(hostport)
}

// MarshalJSON returns the JSON encoding of the node.
func (n *Node) MarshalJSON() ([]byte, error) {
	utcloc, _ := time.LoadLocation("UTC")
	savedata := nodedata{
		ClusterName: n.clusterName,
		Name:        n.name,
		CreatedAt:   n.createdAt.In(utcloc),
		Type:        n.nodetype,
		Ports:       n.ports,
	}

	return json.Marshal(savedata)
}

// UnmarshalJSON  parses and restores a JSON-encoded
// node.
func (n *Node) UnmarshalJSON(b []byte) error {
	var loaddata nodedata

	err := json.Unmarshal(b, &loaddata)
	if err != nil {
		return err
	}

	localloc, _ := time.LoadLocation("Local")

	n.clusterName = loaddata.ClusterName
	n.name = loaddata.Name
	n.createdAt = loaddata.CreatedAt.In(localloc)
	n.nodetype = loaddata.Type
	n.ports = loaddata.Ports

	return nil
}

func (n *Node) createhost() error {
	c := n.Cluster()

	driverimage, err := c.driver.GetImage(c.k8sVersion)
	if err != nil {
		return err
	}

	if driverimage.Deprecated() {
		return errVersionDeprecated
	}

	host, err := c.driver.NewMachine(n.name, c.name, c.k8sVersion)
	if err != nil {
		n.host = nil
		return err
	}
	n.host = host
	return nil
}

func (n *Node) ensurehost() error {
	if n.host == nil {
		c := n.Cluster()
		host, err := c.driver.GetMachine(n.name, c.name)
		if err != nil {
			return err
		}

		n.host = host
	}
	return nil
}
