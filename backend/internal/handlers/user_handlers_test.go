package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stock-in-order/backend/internal/models"
)

// mockUserStore implements userInserter without hitting a real DB.
type mockUserStore struct{}

func (m *mockUserStore) Insert(u *models.User) error {
	// Simulate DB side effects
	u.ID = 1
	u.CreatedAt = u.CreatedAt // leave zero or set later if needed
	return nil
}

func TestRegisterUser_Success(t *testing.T) {
	store := &mockUserStore{}
	h := registerUserHandler(store)

	// Prepare request body
	body := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password123",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Call handler directly
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if resp["email"] != body["email"] {
		t.Fatalf("expected email %s, got %v", body["email"], resp["email"])
	}
}
