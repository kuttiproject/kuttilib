package kuttilib_test

import (
	"testing"

	"github.com/kuttiproject/kuttilib"
	"github.com/kuttiproject/workspace"
)

func testworkspace(t *testing.T) {
	confdir, err := workspace.Configdir()
	if err != nil {
		t.Errorf("getting config dir failed with:%v", err)
	}
	t.Logf("Config dir is :%v", confdir)

	cachedir, err := workspace.Cachedir()
	if err != nil {
		t.Errorf("getting cache dir failed with:%v", err)
	}
	t.Logf("Cache dir is :%v", cachedir)
}

func TestSetWorkspace(t *testing.T) {
	err := kuttilib.SetWorkspace("./out")
	if err != nil {
		t.Errorf("workspace setting failed with error:%v", err)
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

func TestClusterCreation(t *testing.T) {
	testworkspace(t)

	clustercount := len(kuttilib.ClusterNames())
	if clustercount > 0 {
		t.Errorf("there should not be any clusters at this point, but:%v", clustercount)
	}
}
