package installtool

const (
	GoAPPName    = "go"
	GoEnvAPPName = "goenv"
)

var (
	appInstallers = map[string]APPInstaller{}
)

type APPInstaller interface {
	Install(version string) error
	InstallForce(version string) error
	Uninstall(version string) error
	Exists(version string) bool
	NeedRoot() bool
}

func AddInstaller(appName string, installer APPInstaller) {
	appInstallers[appName] = installer
}
