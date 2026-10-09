package local_registry

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type AssetPath struct {
	Provider      string `json:"provider"`
	ArtifactPath  string `json:"artifact_path"`
	MetadataPath  string `json:"metadata_path"`
	ChecksumsPath string `json:"checksums_path"`
}
type AssetContents struct {
	Provider      string `json:"provider"`
	ArtifactJson  string `json:"artifacts_json"`
	MetadataJson  string `json:"metadata_json"`
	ChecksumsText string `json:"checksums_text"`
}

type Asset struct {
}

func (a *Asset) FindAssetPaths(rootDir string) ([]AssetPath, error) {
	byDir := make(map[string]*AssetPath)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		switch {
		case d.Name() == "artifacts.json", d.Name() == "metadata.json":
			dir := filepath.Dir(path)
			entry, ok := byDir[dir]
			if !ok {
				entry = &AssetPath{Provider: filepath.Base(dir)}
				byDir[dir] = entry
			}
			if d.Name() == "artifacts.json" {
				entry.ArtifactPath = path
			}
			if d.Name() == "metadata.json" {
				entry.MetadataPath = path
			}

		case strings.HasSuffix(strings.ToLower(d.Name()), "_checksums.txt"):
			dir := filepath.Dir(path)
			entry, ok := byDir[dir]
			if !ok {
				entry = &AssetPath{Provider: filepath.Base(dir)}
				byDir[dir] = entry
			}
			entry.ChecksumsPath = path
		}

		if strings.Contains(d.Name(), "_Checksums.txt") {
			dir := filepath.Dir(path)
			entry, ok := byDir[dir]
			if !ok {
				entry = &AssetPath{Provider: filepath.Base(dir)}
				byDir[dir] = entry
			}
			entry.ChecksumsPath = path
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Asset-Path: %w", err)
	}

	out := make([]AssetPath, 0, len(byDir))
	for _, v := range byDir {
		out = append(out, *v)
	}

	return out, nil
}

func (a *Asset) FindAssetContents(rootDir string) ([]AssetContents, error) {
	assetPaths, err := a.FindAssetPaths(rootDir)
	if err != nil {
		return nil, err
	}
	var out []AssetContents
	for _, p := range assetPaths {
		var assetcontents AssetContents
		if p.ArtifactPath != "" {
			assetcontents.ArtifactJson = a.readFileContent(p.ArtifactPath)
		}
		if p.MetadataPath != "" {
			assetcontents.MetadataJson = a.readFileContent(p.MetadataPath)
		}
		if p.ChecksumsPath != "" {
			assetcontents.ChecksumsText = a.readFileContent(p.ChecksumsPath)
		}
		assetcontents.Provider = p.Provider
		out = append(out, assetcontents)
	}

	return out, nil
}

func (a *Asset) readFileContent(path string) string {
	content, _ := os.ReadFile(path)
	return string(content)
}
