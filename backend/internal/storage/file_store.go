package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// ErrUserNotFound indicates that the requested user does not exist in the store.
var ErrUserNotFound = errors.New("user not found")

type filePayload struct {
	Users map[string]string `json:"users"`
}

// FileStore provides a file-backed storage for user credentials.
type FileStore struct {
	path  string
	mu    sync.RWMutex
	users map[string]string
}

// NewFileStore constructs a FileStore and ensures the underlying file exists.
func NewFileStore(path string) (*FileStore, error) {
	store := &FileStore{
		path:  path,
		users: make(map[string]string),
	}

	if err := store.ensureFile(); err != nil {
		return nil, err
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// Get retrieves the hashed password for a username.
func (f *FileStore) Get(username string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	hash, ok := f.users[username]
	if !ok {
		return "", ErrUserNotFound
	}
	return hash, nil
}

// Upsert stores or updates the hash for a username and persists the file.
func (f *FileStore) Upsert(username, hash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.users[username] = hash
	return f.persistLocked()
}

func (f *FileStore) ensureFile() error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	if _, err := os.Stat(f.path); errors.Is(err, os.ErrNotExist) {
		initial := filePayload{Users: make(map[string]string)}
		return writeJSONFile(f.path, initial)
	} else if err != nil {
		return err
	}
	return nil
}

func (f *FileStore) load() error {
	file, err := os.Open(f.path)
	if err != nil {
		return err
	}
	defer file.Close()

	var payload filePayload
	if err := json.NewDecoder(file).Decode(&payload); err != nil {
		return err
	}

	if payload.Users == nil {
		payload.Users = make(map[string]string)
	}

	f.users = payload.Users
	return nil
}

func (f *FileStore) persistLocked() error {
	payload := filePayload{
		Users: f.users,
	}
	return writeJSONFile(f.path, payload)
}

func writeJSONFile(path string, payload filePayload) error {
	tempPath := path + ".tmp"

	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(payload); err != nil {
		file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, path)
}
