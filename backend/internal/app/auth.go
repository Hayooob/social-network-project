package app

import (
	"context"
	crypto_rand "crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"social-network/internal/db"
	"social-network/internal/models"
)

const DefaultSessionTTL = 7 * 24 * time.Hour

var (
	ErrEmailAlreadyInUse  = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

// creates a random hex string suitable as a session token.
func generateSessionToken() (string, error) {
	// 32 bytes = 64-char hex
	b := make([]byte, 32)
	if _, err := crypto_rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateUUID() (string, error) {
	b := make([]byte, 16) // 128 bits
	if _, err := crypto_rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// validates input, hashes the password, and inserts a new user.
func RegisterUser(
	ctx context.Context,
	dbConn *sql.DB,
	fullName, dateOfBirth, email, plainPassword string,
	avatarURL, nickname, aboutMe *string,
) (*models.User, error) {
	// check if email already exists
	existing, err := db.GetUserByEmail(dbConn, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyInUse
	}

	if len(plainPassword) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	hash, err := HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	// generate UUID for this user
	uuidStr, err := generateUUID()
	if err != nil {
		return nil, err
	}

	u := &models.User{
		UUID:         uuidStr,
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		DateOfBirth:  dateOfBirth,
		AvatarURL:    avatarURL,
		Nickname:     nickname,
		AboutMe:      aboutMe,
		// IsPrivate default false
	}

	if err := db.CreateUser(dbConn, u); err != nil {
		return nil, err
	}

	return u, nil
}

// checks credentials, creates a session then returns user & session token.
func LoginUser(ctx context.Context, dbConn *sql.DB, email, plainPassword string, ttl time.Duration) (*models.User, string, time.Time, error) {
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}

	u, err := db.GetUserByEmail(dbConn, email)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	if u == nil {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if err := CheckPasswordHash(plainPassword, u.PasswordHash); err != nil {
		// bcrypt returns an error if password doesn't match
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	token, err := generateSessionToken()
	if err != nil {
		return nil, "", time.Time{}, err
	}

	expiresAt := time.Now().Add(ttl)

	if err := db.CreateSession(dbConn, u.ID, token, expiresAt); err != nil {
		return nil, "", time.Time{}, err
	}

	return u, token, expiresAt, nil
}

// returns the user for a session token or nil
func GetUserBySessionToken(ctx context.Context, dbConn *sql.DB, token string) (*models.User, error) {
	sess, err := db.GetSession(dbConn, token)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(sess.ExpiresAt) {
		// delete expired session
		_ = db.DeleteSession(dbConn, token)
		return nil, ErrSessionExpired
	}

	u, err := db.GetUserByID(dbConn, sess.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrSessionNotFound
	}

	return u, nil
}

// deletes the session for the given token.
func LogoutUser(ctx context.Context, dbConn *sql.DB, token string) error {
	return db.DeleteSession(dbConn, token)
}
