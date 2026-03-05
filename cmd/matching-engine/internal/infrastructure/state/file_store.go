package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/domain"
)

type FileStore struct {
	path string
	mu   sync.Mutex
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Load() (map[string]domain.OrderBookSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]domain.OrderBookSnapshot{}, nil
		}
		return nil, err
	}

	var snapshots map[string]domain.OrderBookSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, err
	}

	if snapshots == nil {
		snapshots = map[string]domain.OrderBookSnapshot{}
	}
	return snapshots, nil
}

func (s *FileStore) Save(symbol string, snapshot domain.OrderBookSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := map[string]domain.OrderBookSnapshot{}
	if data, err := os.ReadFile(s.path); err == nil {
		_ = json.Unmarshal(data, &current)
	}

	current[symbol] = snapshot
	return s.writeAtomic(current)
}

func (s *FileStore) writeAtomic(m map[string]domain.OrderBookSnapshot) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
