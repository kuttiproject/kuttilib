package kuttilib_test

import (
	"testing"

	"github.com/kuttiproject/drivercore"
	"github.com/kuttiproject/drivercore/drivercoretest/drivermock"
	"github.com/kuttiproject/kuttilib"
	"github.com/kuttiproject/workspace"
)

const (
	NEWCLUSTER1NAME = "zintakova"
	K8SVERSION1     = "1.23"
	DRIVER1         = "mock1"
	NEWNODE1NAME    = "node1"
	NEWNODE2NAME    = "node2"
	HOSTPORT1       = 10022
	HOSTPORT2       = 20022
)

func init() {
	mock1 := drivermock.New("mock1", "Mock Driver with NAT", true, true)
	if mock1 != nil {
		drivercore.RegisterDriver("mock1", mock1)
		mock1.UpdateRemoteImage(K8SVERSION1, false)
	}
}

func testworkspace(t *testing.T) {
	confdir, err := workspace.ConfigDir()
	if err != nil {
		t.Errorf("getting config dir failed with:%v", err)
	}
	t.Logf("Config dir is :%v", confdir)

	cachedir, err := workspace.CacheDir()
	if err != nil {
		t.Errorf("getting cache dir failed with:%v", err)
	}
	t.Logf("Cache dir is :%v", cachedir)
}

func TestSetWorkspace(t *testing.T) {
	err := kuttilib.SetWorkspace("./out")
	if err != nil {
		t.Fatalf("workspace setting failed with error:%v", err)
	}

	testworkspace(t)
}

func TestValidations(t *testing.T) {
	namevalid := kuttilib.ValidName("validname1")
	if !namevalid {
		t.Error("a valid name was returned as invalid")
	}

	namevalid = kuttilib.ValidName("invalidname")
	if namevalid {
		t.Error("an invalid name was returned as valid")
	}

	portvalid := kuttilib.ValidPort(100)
	if !portvalid {
		t.Error("a valid port was returned as invalid")
	}

	portvalid = kuttilib.ValidPort(65535)
	if !portvalid {
		t.Error("an invalid port was returned as valid")
	}

	clustervalid := kuttilib.ValidateClusterName("newclust1")
	if clustervalid != nil {
		t.Error("a valid cluster name was returned as invalid")
	}

	clustervalid = kuttilib.ValidateClusterName("invalidnewclust1")
	if clustervalid == nil {
		t.Error("an invalid cluster name was returned as valid")
	}

}

func TestDrivers(t *testing.T) {
	drivercount := len(kuttilib.DriverNames())
	if drivercount < 1 {
		t.Fatalf("no drivers found. Expected one")
	}

	mockdriver, ok := kuttilib.GetDriver(DRIVER1)
	if !ok {
		t.Fatalf("expected driver not found")
	}

	versioncount := len(mockdriver.VersionNames())
	if versioncount > 0 {
		t.Fatal("at least one k8s versions found. Expected none")
	}

	err := mockdriver.UpdateVersionList()
	if err != nil {
		t.Fatalf("UpdateVersionList failed with: %v", err)
	}

	versioncount = len(mockdriver.VersionNames())
	if versioncount < 1 {
		t.Fatal("no k8s versions found. Expected one")
	}

	k8sversion, err := mockdriver.GetVersion(K8SVERSION1)
	if err != nil {
		t.Fatalf("getting k8s version %v failed with: %v", K8SVERSION1, err)
	}

	if k8sversion.Status() != kuttilib.VersionStatusNotDownloaded {
		t.Fatalf("k8s version %v should have status NotDownloaded", K8SVERSION1)
	}
}

func TestClusters(t *testing.T) {
	testworkspace(t)

	clustercount := len(kuttilib.ClusterNames())
	if clustercount > 0 {
		t.Errorf("there should not be any clusters at this point, but found %v", clustercount)
	}

	err := kuttilib.NewEmptyCluster(NEWCLUSTER1NAME, K8SVERSION1, DRIVER1)
	if err == nil {
		t.Fatal("expected error in cluster creation. Did not happen")
	}

	mock1, _ := kuttilib.GetDriver(DRIVER1)
	desiredversion, _ := mock1.GetVersion(K8SVERSION1)

	err = desiredversion.Fetch()
	if err != nil {
		t.Fatalf("desired version fetch failed with: %v", err)
	}

	err = kuttilib.NewEmptyCluster(NEWCLUSTER1NAME, K8SVERSION1, DRIVER1)
	if err != nil {
		t.Fatalf("cluster creation failed with error: %v", err)
	}

	_, ok := kuttilib.GetCluster(NEWCLUSTER1NAME)
	if !ok {
		t.Fatalf("new cluster could not be retrieved")
	}

	err = kuttilib.NewEmptyCluster(NEWCLUSTER1NAME, K8SVERSION1, DRIVER1)
	if err == nil {
		t.Fatal("creation of second cluster with same name should have failed. Didn't")
	}

	if len(kuttilib.Clusters()) != 1 {
		t.Fatalf("new cluster cannot be seen in Clusters collection")
	}

	t.Run("TestNodes", testNodes)

	err = kuttilib.DeleteCluster(NEWCLUSTER1NAME, false)
	if err != nil {
		t.Fatalf("cluster delete failed with: %v", err)
	}
}

func testNodes(t *testing.T) {
	cluster, ok := kuttilib.GetCluster(NEWCLUSTER1NAME)
	if !ok {
		t.Fatalf("could not get new cluster")
	}

	node, err := cluster.NewUninitializedNode(NEWNODE1NAME)
	if err != nil {
		t.Fatalf("new node creation failed with: %v", err)
	}

	if node.Status() != kuttilib.NodeStatusStopped {
		t.Fatal("a new node should be in stopped state")
	}

	_, err = cluster.NewUninitializedNode(NEWNODE1NAME)
	if err == nil {
		t.Fatal("second node creation with same name should have failed. Didn't")
	}

	node2, err := cluster.NewUninitializedNode(NEWNODE2NAME)
	if err != nil {
		t.Fatalf("second new node creation failed with: %v", err)
	}

	err = node.ForwardSSHPort(HOSTPORT1)
	if err != nil {
		t.Fatalf("forwarding SSH port failed with: %v", err)
	}

	err = node.CheckHostPort(HOSTPORT2)
	if err != nil {
		t.Fatal("host port should have been occupied. Wasn't")
	}

	err = node.ForwardPort(HOSTPORT1, 22)
	if err == nil {
		t.Fatal("forwardport should have failed because host port occupied. Didn't")
	}

	err = node.ForwardPort(HOSTPORT2, 22)
	if err == nil {
		t.Fatal("forwardport should have failed because node port occupied. Didn't")
	}

	err = node.ForwardPort(HOSTPORT2, 80)
	if err != nil {
		t.Fatalf("Forwardport failed witj: %v", err)
	}

	if portcount := len(node.Ports()); portcount != 2 {
		t.Fatalf("forwarded ports show as %v instead of 2", portcount)
	}

	err = kuttilib.DeleteCluster(NEWCLUSTER1NAME, true)
	if err == nil {
		t.Fatal("cluster delete should have failed with node present. Didn't")
	}

	// Node start etc here
	err = node.Start()
	if err != nil {
		t.Fatalf("node start failed with: %v", err)
	}

	if nodestatus := node.Status(); nodestatus != kuttilib.NodeStatusRunning {
		t.Fatalf("node status is %v instead of running", nodestatus)
	}

	err = cluster.DeleteNode(NEWNODE1NAME, false)
	if err == nil {
		t.Fatal("node delete should have failed. Didn't")
	}

	err = node.Stop()
	if err != nil {
		t.Fatalf("node stop failed with: %v", err)
	}

	err = cluster.DeleteNode(NEWNODE1NAME, false)
	if err != nil {
		t.Fatalf("node delete failed with: %v", err)
	}

	err = node2.ForceStart()
	if err != nil {
		t.Fatalf("node force start failed with: %v", err)
	}

	err = cluster.DeleteNode(NEWNODE2NAME, true)
	if err != nil {
		t.Fatalf("node force delete failed with: %v", err)
	}
}
