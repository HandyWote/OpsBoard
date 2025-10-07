package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/storage"
)

func TestLoginHandler_NewUserCreatesAccount(t *testing.T) {
	store := newMemoryCredentialStore()
	handler := NewLoginHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"alice","password":"secret123"}`))
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	body := decodeBody(t, resp)
	if !body.Success {
		t.Fatalf("expected success true, got false (message: %s)", body.Message)
	}

	cred := store.mustCredential(t, "alice")
	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte("secret123")); err != nil {
		t.Fatalf("stored password hash does not match input: %v", err)
	}

	if !store.loginRecorded(cred.UserID) {
		t.Fatalf("expected login recorded for new user")
	}
}

func TestLoginHandler_ExistingUserValidatesPassword(t *testing.T) {
	store := newMemoryCredentialStore()
	if _, err := store.CreateUserWithPassword(context.Background(), "bob", hashPassword(t, "strongPass!"), hashAlgorithmID, bcryptCost); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	handler := NewLoginHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"bob","password":"strongPass!"}`))
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	body := decodeBody(t, resp)
	if !body.Success {
		t.Fatalf("expected success true, got false (message: %s)", body.Message)
	}
}

func TestLoginHandler_InvalidPassword(t *testing.T) {
	store := newMemoryCredentialStore()
	if _, err := store.CreateUserWithPassword(context.Background(), "carol", hashPassword(t, "secret"), hashAlgorithmID, bcryptCost); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	handler := NewLoginHandler(store)
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"carol","password":"wrong"}`))
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.Code)
	}

	body := decodeBody(t, resp)
	if body.Success {
		t.Fatalf("expected success false, got true")
	}
}

func TestLoginHandler_RejectsBadPayload(t *testing.T) {
	store := newMemoryCredentialStore()
	handler := NewLoginHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":" "}`))
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.Code)
	}

	body := decodeBody(t, resp)
	if body.Success {
		t.Fatalf("expected validation error")
	}
}

func TestLoginHandler_OptionsPreflight(t *testing.T) {
	store := newMemoryCredentialStore()
	handler := NewLoginHandler(store)

	req := httptest.NewRequest(http.MethodOptions, "/api/login", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 for OPTIONS, got %d", resp.Code)
	}
}

type memoryCredentialStore struct {
	mu          sync.RWMutex
	credentials map[string]storage.Credential
	loginCounts map[uuid.UUID]int
}

func newMemoryCredentialStore() *memoryCredentialStore {
	return &memoryCredentialStore{
		credentials: make(map[string]storage.Credential),
		loginCounts: make(map[uuid.UUID]int),
	}
}

func (m *memoryCredentialStore) GetCredential(_ context.Context, username string) (storage.Credential, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cred, ok := m.credentials[strings.ToLower(username)]
	if !ok {
		return storage.Credential{}, storage.ErrUserNotFound
	}
	return cred, nil
}

func (m *memoryCredentialStore) CreateUserWithPassword(_ context.Context, username, passwordHash, algorithm string, cost int) (storage.Credential, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(username))
	if _, exists := m.credentials[key]; exists {
		return storage.Credential{}, errors.New("user already exists")
	}

	cred := storage.Credential{
		UserID:        uuid.New(),
		Username:      username,
		PasswordHash:  passwordHash,
		HashAlgorithm: algorithm,
		HashCost:      cost,
	}

	m.credentials[key] = cred
	return cred, nil
}

func (m *memoryCredentialStore) RecordSuccessfulLogin(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.loginCounts[userID]++
	return nil
}

func (m *memoryCredentialStore) mustCredential(t *testing.T, username string) storage.Credential {
	t.Helper()
	cred, err := m.GetCredential(context.Background(), username)
	if err != nil {
		t.Fatalf("failed to fetch credential for %s: %v", username, err)
	}
	return cred
}

func (m *memoryCredentialStore) loginRecorded(userID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loginCounts[userID] > 0
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

type responseBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func decodeBody(t *testing.T, resp *httptest.ResponseRecorder) responseBody {
	t.Helper()
	var body responseBody
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return body
}
