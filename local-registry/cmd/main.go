package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	localregistry "sonarbridge-go/local-registry"
	"strings"
	"uuid"
)

type ReleaseEntry struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		Url        string `json:"url"`
		IsChecksum bool   `json:"is_checksum"`
	} `json:"assets"`
}

func main() {
	pwd, _ := os.Getwd()
	var port int
	var addr string
	absAssetPath := flag.String("asset-path", pwd, "local-reg [--asset-path]")
	flag.IntVar(&port, "port", 8077, "local-reg [--asset-path]")
	flag.StringVar(&addr, "address", "localhost", "local-reg [--asset-path]")
	flag.Parse()

	server := localregistry.NewLocalRegistryServer(port, addr, *absAssetPath)

	// curl --request GET localhost:8077/assets/binary/vcs-gitlab-latest -o  bitlab.bin
	server.AddFileRoute("/assets/binary/{filename}", "GET", "filename")
	server.AddFileRoute("/assets/checksum/{filename}", "GET", "filename")
	server.AddRoute("/releases/{version}", "POST", func(pvs ...string) any {
		version := pvs[0]
		dirs, err := os.ReadDir(fmt.Sprintf("%s/assets/binaries/%s", *absAssetPath, version))
		if err != nil {
			panic(err)
		}
		releases := ReleaseEntry{TagName: version}
		for _, f := range dirs {
			if f.IsDir() {
				continue
			}
			releases.Assets = append(releases.Assets, struct {
				Id         string `json:"id"`
				Name       string `json:"name"`
				Url        string `json:"url"`
				IsChecksum bool   `json:"is_checksum"`
			}{
				Id:         uuid.New().String(),
				Name:       f.Name(),
				Url:        fmt.Sprintf("%s/assets/binaries/%s/%s", *absAssetPath, version, f.Name()),
				IsChecksum: strings.Contains(f.Name(), ".checksum.txt"),
			})
		}
		return releases
	})
	server.AddRoute("/artifacts/{version}", "GET", func(pvs ...string) any {
		version := pvs[0]

		artifacts := ""
		if version == "" || version == "latest" {
			artifacts = "artifacts.json"
		} else {
			artifacts = version + "/artifacts.json"
		}
		file, _ := os.ReadFile(fmt.Sprintf("%s/assets/%s", *absAssetPath, artifacts))
		var res any
		if err := json.Unmarshal(file, &res); err != nil {
			panic(err)
		}
		return res
	})
	server.Serve()
}
