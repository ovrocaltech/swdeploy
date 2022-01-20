package swdeploy

type Deployer interface {
	Deploy(ver string) error
}
