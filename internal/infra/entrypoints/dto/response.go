package dto

type CreateOrUpdateCommentNoteResponse struct {
}

type Checksum struct {
	BinaryName string `json:"binary_name"`
	Hash       string `json:"hash"`
}
type AssetArtifactResponse struct {
	Provider  string           `json:"provider"`
	Artifacts []map[string]any `json:"artifacts"`
	Metadata  map[string]any   `json:"metadata"`
	Checksums []Checksum       `json:"checksums"`
}

type InstalledPluginStateResponse struct {
	Name            string   `json:"name"`
	State           string   `json:"state"`
	ProtocolVersion string   `json:"protocol_version"`
	Capabilities    []string `json:"capabilities"`
	Type            string   `json:"type"`
	PluginVersion   string   `json:"plugin_version"`
}

type PluginActionResponse struct {
	Name         string   `json:"name"`
	State        string   `json:"state"`
	Changed      bool     `json:"changed"`
	Version      string   `json:"version,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type PluginErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	State   string `json:"state,omitempty"`
}
