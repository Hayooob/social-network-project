package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"social-network/internal/db"
	"social-network/internal/models"
)

// migrationsDir is the schema used by the tests. Pointing the tests at the real
// migrations means a migration that breaks the app also breaks the suite.
const migrationsDir = "../db/migrations"

// testEnv bundles everything a test needs: the server under test, an HTTP
// server in front of it with the real middleware chain, and the database.
type testEnv struct {
	t      *testing.T
	Server *Server
	HTTP   *httptest.Server
}

// newTestEnv builds a server backed by a throwaway SQLite database in the
// test's temporary directory. Each test gets its own database, so tests do not
// share state and can run in any order.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	database := db.OpenDb(dbPath)
	t.Cleanup(func() { _ = database.Close() })

	db.RunMigrations(database, migrationsDir)

	srv := NewServer(database)

	// Same wrapping as Server.Listen, so tests exercise the middleware too.
	handler := srv.CORSMiddleware(srv.AuthMiddleware(srv.Mux))
	httpSrv := httptest.NewServer(handler)
	t.Cleanup(httpSrv.Close)

	return &testEnv{t: t, Server: srv, HTTP: httpSrv}
}

// testUser is a registered account plus an HTTP client that carries its session
// cookie, which is how the real frontend authenticates.
type testUser struct {
	ID       int64
	Email    string
	Password string
	FullName string
	Client   *http.Client
}

// registerUser creates an account directly through the auth layer and logs it
// in over HTTP so the returned client holds a valid session cookie.
func (e *testEnv) registerUser(email, fullName string) *testUser {
	e.t.Helper()

	const password = "password123"

	user, err := RegisterUser(
		context.Background(),
		e.Server.DB,
		fullName,
		"2000-01-01",
		email,
		password,
		nil, nil, nil,
	)
	if err != nil {
		e.t.Fatalf("registerUser(%s): %v", email, err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		e.t.Fatalf("cookiejar.New: %v", err)
	}
	client := &http.Client{Jar: jar}

	tu := &testUser{
		ID:       user.ID,
		Email:    email,
		Password: password,
		FullName: fullName,
		Client:   client,
	}

	resp := e.do(tu, http.MethodPost, "/api/login", map[string]string{
		"email":    email,
		"password": password,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e.t.Fatalf("login for %s: got status %d, want 200", email, resp.StatusCode)
	}

	return tu
}

// setPrivate flips a user's profile to private.
func (e *testEnv) setPrivate(u *testUser, private bool) {
	e.t.Helper()
	if err := db.UpdateUserPrivacy(e.Server.DB, u.ID, private); err != nil {
		e.t.Fatalf("UpdateUserPrivacy(%d, %v): %v", u.ID, private, err)
	}
}

// follow makes follower follow target with the given status, bypassing the HTTP
// layer so tests can set up relationships without depending on handler
// behaviour they are not testing.
func (e *testEnv) follow(follower, target *testUser, status string) {
	e.t.Helper()
	if err := db.CreateFollow(e.Server.DB, int(follower.ID), int(target.ID), status); err != nil {
		e.t.Fatalf("CreateFollow(%d -> %d, %s): %v", follower.ID, target.ID, status, err)
	}
}

// createPost inserts a post for a user with the given privacy.
func (e *testEnv) createPost(author *testUser, content, privacy string, allowedViewers ...int) *models.Post {
	e.t.Helper()

	post, err := db.InsertPostWithExtras(
		e.Server.DB,
		int(author.ID),
		content,
		privacy,
		"",
		allowedViewers,
	)
	if err != nil {
		e.t.Fatalf("InsertPostWithExtras(%s): %v", privacy, err)
	}
	return post
}

// do performs a request against the test server. A nil user sends the request
// without a session cookie (anonymous); a nil body sends no body.
func (e *testEnv) do(u *testUser, method, path string, body any) *http.Response {
	e.t.Helper()

	var reader *bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			e.t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, e.HTTP.URL+path, reader)
	if err != nil {
		e.t.Fatalf("new request %s %s: %v", method, path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := e.HTTP.Client()
	if u != nil {
		client = u.Client
	}

	resp, err := client.Do(req)
	if err != nil {
		e.t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// decodeJSON reads a JSON response body into dst and closes the body.
func decodeJSON(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

// assertStatus fails the test when the response status is not want.
func assertStatus(t *testing.T, resp *http.Response, want int, context string) {
	t.Helper()
	if resp.StatusCode != want {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("%s: got status %d, want %d (body: %v)", context, resp.StatusCode, want, body)
	}
}
