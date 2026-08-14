package cli

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

type Info struct {
	Version  string
	Commit   string
	Date     string
	Dirty    bool
	GoVer    string
	Platform string
}

func Get() Info {
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

func (i Info) ToString() string {
	v := i.Version
	if i.Dirty {
		v += "-salle"
	}
	return fmt.Sprintf("sonarbridge-cli %s (%s, built %s) %s %s", v, i.Commit, i.Date, i.GoVer, i.Platform)
}

func NewVersionCommand() *cobra.Command {
	info := Get()
	return &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Commit: %s\n", Commit)
			fmt.Printf("Built: %s\n", Date)
			fmt.Println(info.ToString())
		},
	}
}
