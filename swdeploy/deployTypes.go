package swdeploy

import (
	"log"
)

type DeployCmd struct {
	MyreposCfg string `yaml:"myrepos_config"`
	// key = host type: ie. gpu, mcs etc
	Cmd map[string]DeployTypes `yaml:"cmd"`
}

type Repo struct {
	Name string
}

//type DeployTypes map[string]DeployUnit `yaml:"shell_repo"`

type DeployTypes struct {
	// key = shell-repo path
	ShellRepo map[string]DeployUnits `yaml:"shell_repo"`
}

// key is 'repos' and or 'services'
type DeployUnits map[string][]string

// write to etcd /mon/deploy
type DeployMonitorData struct {
	Time        float64 `json:"time"`
	Hostname    string  `json:"hostname"`
	Status      string  `json:"status"` // Deploying, Success | Failed
	Error       string  `json:"error"`
	DeployedVer string  `json:"deployed_version"`
	PreVer      string  `json:"previous_version"`
}

/*
func getPrevVersion(fn string) (string, error) {
	return "1.1.1", nil
}

func (cl Repo) Deploy(ver string) error {
	log.Println("Deploying", cl.Name)
	log.Println("  Changing dir to: ", cl.Name)
	log.Println("  Running Deploy for version: ", ver)
	log.Println("  None")
	return nil
}
*/
