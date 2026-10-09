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

type arguments struct {
	port     int
	addr     string
	absAsset string
}

var args arguments

func main() {
	pwd, _ := os.Getwd()
	flag.StringVar(&args.absAsset, "asset-path", pwd, "local-reg [--asset-path]")
	flag.IntVar(&args.port, "port", 8077, "local-reg [--asset-path]")
	flag.StringVar(&args.addr, "address", "localhost", "local-reg [--asset-path]")
	flag.Parse()

	server := localregistry.NewLocalRegistryServer(args.port, args.addr, args.absAsset)

	// curl --request GET localhost:8077/assets/binary/vcs-gitlab-latest -o  bitlab.bin
	server.AddFileRoute("/assets/binary/{filePath}", "GET", "filePath")
	server.AddFileRoute("/assets/checksum/{filename}", "GET", "filename")
	server.AddRoute("/releases/{version}", "POST", func(pvs ...string) any {
		version := pvs[0]
		dirs, err := os.ReadDir(fmt.Sprintf("%s/assets/binaries/gitlab/%s", args.absAsset, version))
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
				Url:        fmt.Sprintf("%s/assets/binaries/%s/%s", args.absAsset, version, f.Name()),
				IsChecksum: strings.Contains(f.Name(), ".checksum.txt"),
			})
		}
		return releases
	})
	server.AddRoute("/{vcs}/artifacts", "GET", fnGetArtifacts)
	server.AddRoute("/{vcs}/artifacts/{version}", "GET", fnGetArtifacts)
	server.AddRoute("/artifacts/metadatas", "GET", fnGetArtifacts)
	server.AddRoute("/artifacts/all", "GET", fnGetAllArtifacts)
	server.Serve()
}

func fnGetAllArtifacts(vps ...string) any {
	dir := fmt.Sprintf("%s/bin.local/assets", args.absAsset)

	asset := &localregistry.Asset{}

	found, err := asset.FindAssetContents(dir)
	if err != nil {
		panic(err)
	}

	return found
}

func fnGetArtifacts(pvs ...string) any {
	var version string
	var vcs string
	if len(pvs) > 0 {
		vcs = pvs[0]
		version = pvs[1]
	}

	artifacts := ""
	if version == "" || version == "latest" {
		artifacts = "artifacts.json"
	} else {
		artifacts = version + "/artifacts.json"
	}
	file, _ := os.ReadFile(fmt.Sprintf("%s/bin.local/assets/binaries/%s/%s", args.absAsset, vcs, artifacts))
	var res any
	if err := json.Unmarshal(file, &res); err != nil {
		panic(err)
	}
	return res
}
