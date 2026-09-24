package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Extra struct {
	Binary     string   `json:"Binary" yaml:"Binary"`
	Binaries   []string `json:"Binaries,omitempty" yaml:"Binaries,omitempty"`
	Files      []string `json:"Files,omitempty" yaml:"Files,omitempty"`
	Builder    string   `json:"Builder" yaml:"Builder"`
	Ext        string   `json:"Ext,omitempty" yaml:"Ext,omitempty"`
	Id         string   `json:"ID" yaml:"ID"`
	Format     string   `json:"Format,omitempty" yaml:"Format,omitempty"`
	ChecksumOf string   `json:"ChecksumOf,omitempty" yaml:"ChecksumOf,omitempty"`
	WrappedIn  string   `json:"WrappedIn,omitempty" yaml:"WrappedIn,omitempty"`
}

// Entry is n-one plugin in the catalog
type Entry struct {
	Name         string `json:"name" yaml:"name"`
	Binary       string `json:"binary,omitempty" yaml:"binary,omitempty"`
	Path         string `json:"path" yaml:"path"`
	InternalType int32  `json:"internal_type" yaml:"internal_type" yaml:"internalType"`
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
	Goos         string `json:"goos,omitempty" yaml:"goos,omitempty"`
	Go386        string `json:"go386,omitempty" yaml:"go386,omitempty"`
	Goarch       string `json:"goarch,omitempty" yaml:"goarch,omitempty"`
	Target       string `json:"target,omitempty" yaml:"target,omitempty"`
	Extra        *Extra `json:"extra,omitempty" yaml:"extra,omitempty"`
}

// Catalog is the parsed content of registry.yaml or registry.json
type Catalog struct {
	Metadata Entry   `yaml:"metadata" json:"metadata"`
	Plugins  []Entry `yaml:"plugins" json:"plugins"`
}

func LoadCatalog(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("registry: read %s: %w", path, err)
	}

	var cat Catalog
	var entries []Entry
	if strings.HasSuffix(path, ".yaml") {
		if err := yaml.Unmarshal(data, &entries); err != nil {
			return nil, fmt.Errorf("registry: parse %s: %w", path, err)
		}
	} else {
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, fmt.Errorf("registry: parse %s: %w", path, err)
		}
	}

	for i, e := range entries {
		if e.Name == "" {
			return nil, fmt.Errorf("registry: %s: entry %d has no name", path, i)
		}

		if e.Type == "Binary" && e.Binary == "" {
			return nil, fmt.Errorf("registry: %s: entry %q has no binary", path, e.Name)
		}

		if e.Type == "Metadata" {
			cat.Metadata = e
		} else {
			cat.Plugins = append(cat.Plugins, e)
		}
	}
	return &cat, nil
}
