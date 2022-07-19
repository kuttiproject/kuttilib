package kuttilib

import (
	"encoding/json"

	"github.com/kuttiproject/workspace"
)

const configFileName = "kuttilib-clusters.json"

var (
	clusterconfigmanager workspace.ConfigManager
	config               *clusterConfigData
)

type clusterConfigData struct {
	Clusters map[string]*Cluster
}

func (cc *clusterConfigData) Serialize() ([]byte, error) {
	return json.Marshal(cc)
}

func (cc *clusterConfigData) Deserialize(data []byte) error {
	var loadedconfig *clusterConfigData
	err := json.Unmarshal(data, &loadedconfig)
	if err == nil {
		cc.Clusters = loadedconfig.Clusters
	}

	return err
}

func (cc *clusterConfigData) SetDefaults() {
	cc.Clusters = map[string]*Cluster{}
}

func setworkspaceconfigmanager() {
	config = &clusterConfigData{
		Clusters: map[string]*Cluster{},
	}

	var err error
	clusterconfigmanager, err = workspace.NewFileConfigManager(configFileName, config)
	if err != nil {
		panic("could not initialize cluster configuration manager")
	}
}

func init() {
	setworkspaceconfigmanager()
}
