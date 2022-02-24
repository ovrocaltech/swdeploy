// deploy is a service which listens for the command(s) sepecified in the
// deployCfg.yml file and the payload will be the version of the shell repo
// to checkout. It will checkout the specified shell version and
// change directory into each of the defined 'repo' listed and run the deploy
// script. When all repos have been deployed, it will restart each service
// in the 'service' list using 'systemctl restart <service>'.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	cmddxformat "git.ovro.caltech.edu/sw/git/cmddxformat.git"
	du "git.ovro.caltech.edu/sw/git/dsa110-goutils.git"
	etcdaccess "git.ovro.caltech.edu/sw/git/etcdaccess.git/v2"
	gw "git.ovro.caltech.edu/sw/git/gogitwrapper.git"
	dt "github.com/ovrocaltech/swdeploy.git/swdeploy"
	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/soniakeys/meeus/v3/julian"
	"log/syslog"
	"os"
	"os/exec"
	"time"
)

const (
	serverConfigFilename = "deployCfg.yml"
	ETCD_DEPLOY_KEY      = "/cmd/deploy"
	// monitor data will be written out under ETCD_MON_KEY/<hostname>
	ETCD_MON_KEY = "/mon/deploy"
	MYREPO       = "mr"
	NOMATCH      = "noMatch"
	MJDOFFSET    = 2400000.5
)

var (
	// Version is the git version at build time
	Version string
	// Build is git hash at buid time
	Build   string
	ctxlog  *log.Entry
	appName string
	srvCfg  dt.DeployCmd
)

func init() {
	appName := "deploy"
	log.SetFormatter(&log.JSONFormatter{})

	hook, err := logrus_syslog.NewSyslogHook("tcp", "localhost:514", syslog.LOG_INFO, "")
	if err != nil {
		log.Error("Unable to connect to local syslog daemon")
	} else {
		log.AddHook(hook)
	}
	standardFields := log.Fields{
		"Version": Version,
		"Build":   Build,
		"app":     appName,
	}
	ctxlog = log.WithFields(standardFields)

	err = du.ReadYaml(serverConfigFilename, &srvCfg)
	if err != nil {
		emsg := fmt.Sprintf("Error reading %s\n", serverConfigFilename)
		ctxlog.Error(emsg)
		panic(0)
	}
}

func writeMonitorData(md dt.DeployMonitorData) {
	md.Time = getMJD()

	j, err := json.Marshal(md)
	if err != nil {
		ctxlog.Error("Unable to Marshal monitor data: ", err.Error())
		return
	}
	etcdaccess.Put(ETCD_MON_KEY+"/"+md.Hostname, string(j))

	j, _ = json.Marshal(md)
	etcdaccess.Put(ETCD_MON_KEY+"/"+md.Hostname, string(j))
}

func cloneRepo(repo string) error {
	fmt.Println("TODO: clone repo", repo)
	/*
		if ok := repoExists(repo); !ok {
			// clone repo
			// return err
		}
	*/
	return nil
}

func getCurrentVersionSim(repo string) string {
	fmt.Println("TODO: implement getCurrentVersion")
	// get version of repo
	// cd repo
	return "v1.0.0"
}

func restoreRepo() error {
	cmd := exec.Command("git", "checkout", "--", ".")
	_, err := cmd.Output()
	return err
}

// deployRepo runs the deploy script in the repo given by its full path as rp.
func deployRepo(rp string) error {
	log.Println("Changing dir to: ", rp)
	// TODO: fill in monitor data with status DEPLOYING and write out.
	err := os.Chdir(rp)
	if err != nil {
		ctxlog.Error(err.Error())
		return err
	} else {
		err := restoreRepo()
		if err != nil {
			ctxlog.Error(err.Error())
			return err
		}
		//   run deploy script and check for errors
		cmd := exec.Command("./deploy")
		_, err = cmd.Output()
		if err != nil {
			ctxlog.Error(err.Error())
			return err
		}
	}
	return err
}

// checkoutShell executes mr -c .mrconfig_production fetch in specified dir
func checkoutShellRepos(sp string) error {
	ctxlog.Info("checkoutShellRepos...")
	err := os.Chdir(sp)
	if err != nil {
		return err
	}
	cmd := exec.Command(MYREPO, "-c", srvCfg.MyreposCfg, "checkout")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		emsg := fmt.Sprintf("%s %s", err.Error(), out)
		return errors.New(emsg)
	}
	return nil
}

// fetch Shell repos executes mr -c .mrconfig_production fetch in specified dir
func updateShellRepos(sp string) error {
	ctxlog.Info("updateShellRepos...")
	err := os.Chdir(sp)
	cmd := exec.Command(MYREPO, "-c", srvCfg.MyreposCfg, "update")
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	return err
}

func deployCode(repoPath string) error {
	log.Println("Deploying repo: ", repoPath)
	err := deployRepo(repoPath)
	if err != nil {
		ctxlog.Error(err.Error())
		emsg := fmt.Sprintf("Error deploying repo: %s. %s",
			repoPath, err.Error())
		return errors.New(emsg)
	}
	return nil
}
func getMJD() float64 {
	currentTime := time.Now()
	jd := julian.TimeToJD(currentTime.UTC())
	mjd := jd - MJDOFFSET
	return mjd
}

func containsCmd(cmd string, a map[string]dt.DeployTypes) string {
	for k, _ := range a {
		if k == cmd {
			return cmd
		}
	}
	return NOMATCH
}

func fetchShellRepo(shellRepo string, monData *dt.DeployMonitorData) {
	err := gw.FetchRepo(shellRepo)
	if err != nil {
		if err.Error() == "already up-to-date" {
			ctxlog.Info("Repo already up to date")
		} else {
			ctxlog.Error(err.Error())
			monData.Status = "FAILED"
			monData.Error = err.Error()
			writeMonitorData(*monData)
		}
	}
}

// getCurrentVersion returns the current version tag of the repo.
func getCurrentVersion(shellRepo string) string {
	// get current version to later save as prev. version after
	// update
	currentShellVersion, err := gw.GetCurrentVersionName(shellRepo)
	if err != nil {
		ctxlog.Error(err.Error())
	} else {
		ctxlog.Info("Current shell version: ", currentShellVersion)
	}
	return currentShellVersion
}

func checkoutShellAtVersion(shellRepo, currentShellVer, shellVer string,
	monData *dt.DeployMonitorData) error {

	ctxlog.Info("calling gw.CheckoutTag")
	err := gw.CheckoutTag(shellRepo, shellVer)
	if err != nil {
		ctxlog.Error(err.Error())
	} else {
		monData.PreVer = currentShellVer
		monData.DeployedVer = shellVer
	}
	return err
}

func deployRepos(shellRepo string, repos []string, monData *dt.DeployMonitorData) {
	for _, repo := range repos {
		log.Println("deployRepos in shell: ", shellRepo)
		repoPath := shellRepo + "/" + repo
		err := deployCode(repoPath)
		if err != nil {
			ctxlog.Error(err.Error())
			monData.Status = "FAILED"
			monData.Error = err.Error()
			return
		} else {
			monData.Status = "SUCCESS"
			monData.Error = ""
		}
	}
}

func restartServices(services []string, monData *dt.DeployMonitorData) error {
	for _, srv := range services {
		log.Println("restarting service: ", srv)
		err := restartService(srv)
		if err != nil {
			ctxlog.Error(err.Error())
			monData.Status = "FAILED"
			monData.Error = err.Error()
			return err
		}
	}
	return nil
}

func restartService(srv string) error {
	cmd := exec.Command("sudo", "systemctl", "restart", srv)
	_, err := cmd.Output()
	if err != nil {
		ctxlog.Error(err.Error())
	}
	return err
}

func listenAndServe() {
	rch := etcdaccess.Watch(ETCD_DEPLOY_KEY)
	log.Println("Got etcd")
	var monData dt.DeployMonitorData
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	monData.Hostname = hostname
	log.Printf("Listening for commands...")
	monData.Status = "Listening"
	writeMonitorData(monData)

	for {
		select {
		case wresp := <-rch:
			for _, ev := range wresp.Events {

				cmd := cmddxformat.CommandInterface{}
				//log.Printf("CMD %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				err := json.Unmarshal([]byte(ev.Kv.Value), &cmd)
				if err != nil {
					ctxlog.Error(err.Error())
				}
				log.Println("CMD ", cmd.Cmd)
				switch cmd.Cmd {
				case containsCmd(cmd.Cmd, srvCfg.Cmd):
					shellVersion := cmd.Val.(string)
					ctxlog.Info("desired shell_version= ", shellVersion)
					if err == nil {
						//if err != nil && !strings.Contains(err.Error(), "already") {
						// Note: If no new changes are pulled into repo, then an err
						// is returned with msg: already up to date an no further action
						// is taken

						//functions to implement
						// TODO: if shell doesn't exist, clone it. otherwise return nil.
						// First, update repo to get latest tags
						// for every shell repo, install the list of repos
						for shellRepo, targets := range srvCfg.Cmd[cmd.Cmd].ShellRepo {
							ctxlog.Info("calling gw.FetchRepo")
							monData.Status = fmt.Sprintf("Fetching Shell %s", shellRepo)
							monData.Error = ""
							writeMonitorData(monData)

							fetchShellRepo(shellRepo, &monData)
							writeMonitorData(monData)

							currentShellVersion := getCurrentVersion(shellRepo)

							// checkout shell at version, then perform the actual checkout
							err := checkoutShellAtVersion(shellRepo, currentShellVersion,
								shellVersion, &monData)
							if err != nil {
								monData.Status = "FAILED Shell checkout"
								monData.Error = err.Error()
							} else {
								monData.Status = "Checking out Shell Repos"
								writeMonitorData(monData)
								err := checkoutShellRepos(shellRepo)
								if err != nil {
									monData.Status = "Failed checkout Repos"
									monData.Error = err.Error()
								} else {
									monData.Status = "Updating Shell Repos"
									writeMonitorData(monData)
									err := updateShellRepos(shellRepo)
									if err != nil {
										monData.Status = "Failed Updating Repos"
										monData.Error = err.Error()
									} else {
										monData.Status = "Deploying Repos"
										writeMonitorData(monData)
										deployRepos(shellRepo, targets["repos"], &monData)
									}
								}
							}
							writeMonitorData(monData)
							// if error, discontinue deployment
							if monData.Error != "" {
								break
							}
						} // end of shell repo map loop

						// restart services
						if monData.Error == "" {
							for _, targets := range srvCfg.Cmd[cmd.Cmd].ShellRepo {
								monData.Status = "Restarting Services"
								writeMonitorData(monData)
								err := restartServices(targets["services"], &monData)
								if err == nil {
									monData.Status = "SUCCESS"
								}
								writeMonitorData(monData)
								// if error, discontinue
								if monData.Error != "" {
									break
								}
							} // end of shell repo map loop
						}
					}
				}
			}
		}
	}
}

func main() {
	log.Println("Version: ", Version)
	log.Println("Build: ", Build)

	listenAndServe()

}
