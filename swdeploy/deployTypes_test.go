package swdeploy

import (
	"fmt"
	goutils "git.ovro.caltech.edu/sw/git/dsa110-goutils.git"
	//	"gopkg.in/yaml.v2"
	"testing"
)

func TestGetPreviousVersion(t *testing.T) {
	prevVersion, err := getPrevVersion("testfn")
	if err != nil {
		fmt.Println("error getting previous version: ", err.Error())
		t.Fail()
	}
	fmt.Println("prev. version= ", prevVersion)
}

func deployIt(r Repo, ver string) {
	r.Deploy(ver)
}

func TestConfig(t *testing.T) {

	/*
		var dc2 DeployCmd
		dc.ShellRepo = "/ab/c/"
		var dt DeployTypes
		ra := make([]Repo, 2)
		ra[0].Name = "repo0"
		ra[1].Name = "repo1"
		dt.Repos = ra
		se := make([]string, 2)
		se[0] = "ser1"
		se[1] = "ser2"
		dt.Services = se
		dc.Cmd = make(map[string]DeployTypes)
		dc.Cmd["gpu"] = dt
		dc.Cmd["wx"] = dt
		fmt.Println("dc1", dc)

		dc_yml, err := yaml.Marshal(dc)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("dc1_yml: ", string(dc_yml))
	*/

	var dc2 DeployCmd
	err := goutils.ReadYaml("testDeployCfg.yml", &dc2)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Printf("dc2= %v\n", dc2)

	// check for gpu cmd
	if val, ok := dc2.Cmd["gpu"]; !ok {
		fmt.Println("gpu cmd not found")
		t.Fail()
	} else {
		if unit, ok := val.ShellRepo["/home/ubuntu/proj/lwa-shell"]; !ok {
			fmt.Println("/home/ubuntu/proj/lwa-shell not found ")
			t.Fail()
		} else {
			if repos, ok := unit["repos"]; !ok {
				fmt.Println("repos not found")
				t.Fail()
			} else {
				if repos[0] != "repo1" {
					fmt.Print("Expected repo1, got ", repos[0])
					t.Fail()
				}
				if repos[1] != "repo2" {
					fmt.Print("Expected repo2, got ", repos[1])
					t.Fail()
				}
			}
			// Check services
			if srvs, ok := unit["services"]; !ok {
				fmt.Println("services not found")
				t.Fail()
			} else {
				if srvs[0] != "service1" {
					fmt.Print("Expected service1, got ", srvs[0])
					t.Fail()
				}
				if srvs[1] != "service2" {
					fmt.Print("Expected service2, got ", srvs[1])
					t.Fail()
				}
			}
		}
		// check next shell repo
		if unit, ok := val.ShellRepo["/home/ubuntu/proj/ovrocaltech-shell"]; !ok {
			fmt.Println("/home/ubuntu/proj/ovrocaltech-shell not found ")
			t.Fail()
		} else {
			if repos, ok := unit["repos"]; !ok {
				fmt.Println("repos not found")
				t.Fail()
			} else {
				if repos[0] != "deploy-test-2" {
					fmt.Print("Expected deploy-test-2, got ", repos[0])
					t.Fail()
				}
				if repos[1] != "deploy-test" {
					fmt.Print("Expected deploy_test, got ", repos[1])
					t.Fail()
				}
			}
			// Check services
			if srvs, ok := unit["services"]; !ok {
				fmt.Println("services not found")
				t.Fail()
			} else {
				if srvs[0] != "service-t1" {
					fmt.Print("Expected service-t1, got ", srvs[0])
					t.Fail()
				}
				if srvs[1] != "service-t2" {
					fmt.Print("Expected service-t2, got ", srvs[1])
					t.Fail()
				}
			}
		}
	}
	// check for wx cmd
	if val, ok := dc2.Cmd["wx"]; !ok {
		fmt.Println("wx cmd not found")
		t.Fail()
	} else {
		if unit, ok := val.ShellRepo["/home/ubuntu/proj/dsa110-shell"]; !ok {
			fmt.Println("/home/ubuntu/proj/dsa110-shell not found ")
			t.Fail()
		} else {
			if repos, ok := unit["repos"]; !ok {
				fmt.Println("repos not found")
				t.Fail()
			} else {
				if repos[0] != "repowx" {
					fmt.Print("Expected repowx, got ", repos[0])
					t.Fail()
				}
			}
			// Check services
			if srvs, ok := unit["services"]; !ok {
				fmt.Println("services not found")
				t.Fail()
			} else {
				if srvs[0] != "wx" {
					fmt.Print("Expected wx, got ", srvs[0])
					t.Fail()
				}
			}
		}
	}
}

func TestDeploy(t *testing.T) {
	var apc Repo
	apc.Name = "lwa-shell"

	deployIt(apc, "1.5.0")
}

func TestDeployConfig(t *testing.T) {
	var dc DeployCmd
	err := goutils.ReadYaml("testDeployCfg.yml", &dc)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	for cmd, dt := range dc.Cmd {
		fmt.Println("Deploying cmd: ", cmd)
		for shell, targets := range dt.ShellRepo {
			fmt.Println("Deploying shell: ", shell)
			for key, target := range targets {
				if key == "repos" {
					for _, repo := range target {
						var apc Repo
						apc.Name = repo
						deployIt(apc, "1.5.0")
					}
				}
			}
		}
	}

}

func TestRestartServices(t *testing.T) {
	var dc DeployCmd
	err := goutils.ReadYaml("testDeployCfg.yml", &dc)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	for cmd, dt := range dc.Cmd {
		fmt.Println("Deploying cmd: ", cmd)
		for shell, targets := range dt.ShellRepo {
			fmt.Println("Deploying shell: ", shell)
			for key, target := range targets {
				if key == "services" {
					for _, srv := range target {
						fmt.Println("restarting service: ", srv)
					}
				}
			}
		}
	}
}
