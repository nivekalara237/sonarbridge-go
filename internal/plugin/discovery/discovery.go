package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest is the metadata writen by registry.Install alongside the
// downloaded executable (plugins/<name>/plugin.json)
type Manifest struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Type            string   `json:"type"`
	ProtocolVersion string   `json:"protocolVersion"`
	Capabilities    []string `json:"capabilities"`
	Binary          string   `json:"binary"`
	SHA256          string   `json:"sha256"`
}

// Discovered pairs the manifest with the resolved obsolete path to its
// executable, ready to be handled to runtime.Adapter.Start
type Discovered struct {
	Manifest   Manifest
	BinaryPath string
}

// Scan walks pluginDir (one subdirectory per installed plugin, each contains plugin.json + its executable)
// and returns every valid manifest found.
func Scan(pluginDir string) (found []Discovered, errs []error) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("discovery: read %s: %w", pluginDir, err)}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(pluginDir, entry.Name())
		manifestPath := filepath.Join(dir, "plugin.json")

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("discovery: %s: %w", manifestPath, err))
			continue
		}

		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			errs = append(errs, fmt.Errorf("discovery: parse %s: %w", manifestPath, err))
			continue
		}

		found = append(found, Discovered{
			Manifest:   m,
			BinaryPath: filepath.Join(dir, m.Binary),
		})
	}

	return found, errs
}
