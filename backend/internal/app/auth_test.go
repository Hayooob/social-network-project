package app

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"social-network/internal/db"
)

// --- password hashing -------------------------------------------------------

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if hash == "correct horse battery staple" {
		t.Fatal("password was stored in plain text")
	}

	if err := CheckPasswordHash("correct horse battery staple", hash); err != nil {
		t.Fatalf("CheckPasswordHash with the correct password: %v", err)
	}
}

func TestCheckPasswordHashRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if err := CheckPasswordHash("password124", hash); err == nil {
		t.Fatal("CheckPasswordHash accepted an incorrect password")
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	first, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	second, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if first == second {
		t.Fatal("hashing the same password twice produced identical hashes, so it is not salted")
	}
}

// --- registration -----------------------------------------------------------

func TestRegisterUserStoresHashedPassword(t *testing.T) {
	env := newTestEnv(t)

	user, err := RegisterUser(context.Background(), env.Server.DB,
		"Ada Lovelace", "1815-12-10", "ada@example.com", "password123", nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	if user.PasswordHash == "password123" {
		t.Fatal("password stored in plain text")
	}
	if !strings.HasPrefix(user.PasswordHash, "$2") {
		t.Fatalf("password hash %q does not look like bcrypt", user.PasswordHash)
	}
}

// TestRegisterUserPopulatesGeneratedFields guards a bug where CreateUser
// discarded the result of the INSERT, so the register endpoint answered with
// "id": 0 and a zero timestamp for every new account.
func TestRegisterUserPopulatesGeneratedFields(t *testing.T) {
	env := newTestEnv(t)

	user, err := RegisterUser(context.Background(), env.Server.DB,
		"Grace Hopper", "1906-12-09", "grace@example.com", "password123", nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	if user.ID == 0 {
		t.Error("user ID is 0; the database-generated ID was not read back")
	}
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt is the zero time; it was not read back from the database")
	}
}

func TestRegisterUserRejectsDuplicateEmail(t *testing.T) {
	env := newTestEnv(t)

	if _, err := RegisterUser(context.Background(), env.Server.DB,
		"First User", "2000-01-01", "taken@example.com", "password123", nil, nil, nil); err != nil {
		t.Fatalf("first RegisterUser: %v", err)
	}

	_, err := RegisterUser(context.Background(), env.Server.DB,
		"Second User", "2000-01-01", "taken@example.com", "password123", nil, nil, nil)
	if err != ErrEmailAlreadyInUse {
		t.Fatalf("got error %v, want ErrEmailAlreadyInUse", err)
	}
}

func TestRegisterUserRejectsShortPassword(t *testing.T) {
	env := newTestEnv(t)

	_, err := RegisterUser(context.Background(), env.Server.DB,
		"Short Password", "2000-01-01", "short@example.com", "1234567", nil, nil, nil)
	if err == nil {
		t.Fatal("RegisterUser accepted a 7-character password")
	}
}

func TestRegisterUserGeneratesDistinctUUIDs(t *testing.T) {
	env := newTestEnv(t)

	first, err := RegisterUser(context.Background(), env.Server.DB,
		"User One", "2000-01-01", "one@example.com", "password123", nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}
	second, err := RegisterUser(context.Background(), env.Server.DB,
		"User Two", "2000-01-01", "two@example.com", "password123", nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	if first.UUID == second.UUID {
		t.Fatalf("two users share the UUID %q", first.UUID)
	}
	if first.UUID == "" {
		t.Fatal("UUID is empty")
	}
}

// --- login ------------------------------------------------------------------

func TestLoginUserSucceedsWithCorrectCredentials(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("login@example.com", "Login User")

	got, token, expiresAt, err := LoginUser(context.Background(), env.Server.DB,
		user.Email, user.Password, 0)
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	if got.ID != user.ID {
		t.Errorf("logged in as user %d, want %d", got.ID, user.ID)
	}
	if len(token) != 64 {
		t.Errorf("session token is %d characters, want 64 (32 random bytes, hex encoded)", len(token))
	}
	if !expiresAt.After(time.Now()) {
		t.Error("session expires in the past")
	}
}

func TestLoginUserRejectsWrongPassword(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("wrongpass@example.com", "Wrong Pass")

	_, _, _, err := LoginUser(context.Background(), env.Server.DB, user.Email, "not-the-password", 0)
	if err != ErrInvalidCredentials {
		t.Fatalf("got error %v, want ErrInvalidCredentials", err)
	}
}

// TestLoginUserRejectsUnknownEmailIndistinguishably checks that an unknown
// email and a wrong password fail the same way, so the endpoint cannot be used
// to find out which addresses have accounts.
func TestLoginUserRejectsUnknownEmailIndistinguishably(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("known@example.com", "Known User")

	_, _, _, unknownErr := LoginUser(context.Background(), env.Server.DB, "nobody@example.com", "password123", 0)
	_, _, _, wrongPassErr := LoginUser(context.Background(), env.Server.DB, user.Email, "wrong-password", 0)

	if unknownErr != ErrInvalidCredentials {
		t.Fatalf("unknown email: got %v, want ErrInvalidCredentials", unknownErr)
	}
	if unknownErr != wrongPassErr {
		t.Errorf("unknown email returns %v but wrong password returns %v; these must match", unknownErr, wrongPassErr)
	}
}

func TestLoginUserHonoursCustomTTL(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("ttl@example.com", "TTL User")

	_, _, expiresAt, err := LoginUser(context.Background(), env.Server.DB, user.Email, user.Password, time.Hour)
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	if remaining := time.Until(expiresAt); remaining > time.Hour+time.Minute || remaining < time.Hour-time.Minute {
		t.Errorf("session expires in %v, want roughly 1h", remaining)
	}
}

func TestLoginUserIssuesUniqueTokens(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("unique@example.com", "Unique Tokens")

	_, first, _, err := LoginUser(context.Background(), env.Server.DB, user.Email, user.Password, 0)
	if err != nil {
		t.Fatalf("first LoginUser: %v", err)
	}
	_, second, _, err := LoginUser(context.Background(), env.Server.DB, user.Email, user.Password, 0)
	if err != nil {
		t.Fatalf("second LoginUser: %v", err)
	}

	if first == second {
		t.Fatal("two logins produced the same session token")
	}
}

// --- sessions ---------------------------------------------------------------

func TestGetUserBySessionTokenReturnsUser(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("session@example.com", "Session User")

	_, token, _, err := LoginUser(context.Background(), env.Server.DB, user.Email, user.Password, 0)
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	got, err := GetUserBySessionToken(context.Background(), env.Server.DB, token)
	if err != nil {
		t.Fatalf("GetUserBySessionToken: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("resolved to user %d, want %d", got.ID, user.ID)
	}
}

func TestGetUserBySessionTokenRejectsUnknownToken(t *testing.T) {
	env := newTestEnv(t)

	_, err := GetUserBySessionToken(context.Background(), env.Server.DB, "not-a-real-token")
	if err != ErrSessionNotFound {
		t.Fatalf("got error %v, want ErrSessionNotFound", err)
	}
}

func TestGetUserBySessionTokenRejectsAndClearsExpiredSession(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("expired@example.com", "Expired Session")

	token := "expired-token-for-test"
	if err := db.CreateSession(env.Server.DB, user.ID, token, time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if _, err := GetUserBySessionToken(context.Background(), env.Server.DB, token); err != ErrSessionExpired {
		t.Fatalf("got error %v, want ErrSessionExpired", err)
	}

	// The expired row should have been cleaned up rather than left behind.
	sess, err := db.GetSession(env.Server.DB, token)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if sess != nil {
		t.Error("expired session is still in the database")
	}
}

func TestLogoutUserInvalidatesSession(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("logout@example.com", "Logout User")

	_, token, _, err := LoginUser(context.Background(), env.Server.DB, user.Email, user.Password, 0)
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	if err := LogoutUser(context.Background(), env.Server.DB, token); err != nil {
		t.Fatalf("LogoutUser: %v", err)
	}

	if _, err := GetUserBySessionToken(context.Background(), env.Server.DB, token); err != ErrSessionNotFound {
		t.Fatalf("after logout got error %v, want ErrSessionNotFound", err)
	}
}

// --- HTTP layer -------------------------------------------------------------

func TestRegisterEndpointNeverReturnsPasswordHash(t *testing.T) {
	env := newTestEnv(t)

	resp := env.do(nil, http.MethodPost, "/api/register", map[string]string{
		"email":           "http-register@example.com",
		"password":        "password123",
		"confirmPassword": "password123",
		"first_name":      "HTTP",
		"last_name":       "Register",
		"date_of_birth":   "2000-01-01",
	})
	assertStatus(t, resp, http.StatusCreated, "POST /api/register")

	var body map[string]any
	decodeJSON(t, resp, &body)

	for _, forbidden := range []string{"password", "password_hash", "PasswordHash"} {
		if _, present := body[forbidden]; present {
			t.Errorf("register response exposes %q", forbidden)
		}
	}

	if id, ok := body["id"].(float64); !ok || id == 0 {
		t.Errorf("register response id is %v, want a real database ID", body["id"])
	}
}

func TestRegisterEndpointRejectsMismatchedPasswords(t *testing.T) {
	env := newTestEnv(t)

	resp := env.do(nil, http.MethodPost, "/api/register", map[string]string{
		"email":           "mismatch@example.com",
		"password":        "password123",
		"confirmPassword": "password456",
		"first_name":      "Mis",
		"last_name":       "Match",
		"date_of_birth":   "2000-01-01",
	})
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest, "POST /api/register with mismatched passwords")
}

func TestLoginEndpointSetsHttpOnlySessionCookie(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("cookie@example.com", "Cookie User")

	resp := env.do(nil, http.MethodPost, "/api/login", map[string]string{
		"email":    user.Email,
		"password": user.Password,
	})
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK, "POST /api/login")

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			sessionCookie = c
		}
	}

	if sessionCookie == nil {
		t.Fatal("login did not set a session_id cookie")
	}
	if !sessionCookie.HttpOnly {
		t.Error("session cookie is not HttpOnly, so scripts on the page can read it")
	}
	if sessionCookie.Value == "" {
		t.Error("session cookie is empty")
	}
}

func TestMeEndpointRequiresAuthentication(t *testing.T) {
	env := newTestEnv(t)

	resp := env.do(nil, http.MethodGet, "/api/me", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized, "GET /api/me without a session")
}

func TestMeEndpointReturnsCurrentUser(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("me@example.com", "Me User")

	resp := env.do(user, http.MethodGet, "/api/me", nil)
	assertStatus(t, resp, http.StatusOK, "GET /api/me")

	var body map[string]any
	decodeJSON(t, resp, &body)

	if body["email"] != user.Email {
		t.Errorf("got email %v, want %s", body["email"], user.Email)
	}
	if _, present := body["password_hash"]; present {
		t.Error("/api/me exposes password_hash")
	}
}

func TestLogoutEndpointClearsSession(t *testing.T) {
	env := newTestEnv(t)
	user := env.registerUser("httplogout@example.com", "HTTP Logout")

	resp := env.do(user, http.MethodPost, "/api/logout", nil)
	resp.Body.Close()
	assertStatus(t, resp, http.StatusOK, "POST /api/logout")

	// The same client should now be anonymous.
	after := env.do(user, http.MethodGet, "/api/me", nil)
	defer after.Body.Close()
	assertStatus(t, after, http.StatusUnauthorized, "GET /api/me after logout")
}

func TestAuthMiddlewareIgnoresInvalidSessionCookie(t *testing.T) {
	env := newTestEnv(t)

	req, err := http.NewRequest(http.MethodGet, env.HTTP.URL+"/api/me", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "forged-session-token"})

	resp, err := env.HTTP.Client().Do(req)
	if err != nil {
		t.Fatalf("GET /api/me: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized, "GET /api/me with a forged cookie")
}
