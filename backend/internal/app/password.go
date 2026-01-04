package app

import "golang.org/x/crypto/bcrypt"

// hash a plain text password using bcrypt 
func HashPassword(plain string) (string, error) {
	// bcrypt.DefaultCost is a good default 
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// compare a plain-text password with a stored hash
func CheckPasswordHash(plain, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
