package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type PluginRecord struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

// Store is a small JSON-backed, atomic-write persistence layer. Swapping
// fpr BoltDB later (it concurrent transactional writes become necessary)
// means a new implementation of the same Load/Save shape, not a change
// to callers.
type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}


// Load reads all plugin records, returning by the JSON file at path
// (created on first Save if it doesn't exist)
func (s *Store) Load() (map[string]PluginRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return map[string]PluginRecord{}, nil
	}

	if err != nil {
		return nil, err
	}

	var records map[string]PluginRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// Save write all plugin records atomically (write to a temp file, then rename - safe
// against a crash mid-write on POSIX filesystems).
func (s *Store) Save(records map[string]PluginRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(records, "",  " ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err!=nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
