package build

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

type Info struct {
	Version  string
	Commit   string
	Date     string
	Dirty    bool
	GoVer    string
	Platform string
}

func GetBuildInfo() Info {
	i := Info{
		Version:  Version,
		Commit:   Commit,
		Date:     Date,
		GoVer:    runtime.Version(),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return i
	}

	fmt.Println(&bi)

	if i.Version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		i.Version = bi.Main.Version
	}

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if i.Commit == "" && len(s.Value) >= 7 {
				i.Commit = s.Value[:7]
			}
		case "vcs.time":
			if i.Date == "" {
				i.Date = s.Value
			}
		case "vcs.modified":
			i.Dirty = s.Value == "true"
		}
	}

	return i
}

func (i Info) ToCliString() string {
	v := i.Version
	if i.Dirty {
		v += "-salle"
	}
	return fmt.Sprintf("%s %s (%s, built %s) %s %s", "sonarbridge-cli", v, i.Commit, i.Date, i.GoVer, i.Platform)
}

func (i Info) ToServerString() string {
	v := i.Version
	if i.Dirty {
		v += "-salle"
	}
	return fmt.Sprintf("%s %s (%s, built %s) %s %s", "sonarbridge-server", v, i.Commit, i.Date, i.GoVer, i.Platform)
}
