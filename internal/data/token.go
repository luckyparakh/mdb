package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type TokenModel struct {
	DB *sql.DB
}

func ValidateTokenPlaintext(plainToken string) bool {
	if plainToken == "" || len(plainToken) != 26 {
		return false
	}
	return true
}
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomByte, err := generateRandom()
	if err != nil {
		return nil, err
	}
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomByte)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

func generateRandom() ([]byte, error) {
	randomByte := make([]byte, 16)
	_, err := rand.Read(randomByte)
	if err != nil {
		return nil, err
	}
	return randomByte, nil
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `INSERT INTO tokens (hash, user_id, expiry, scope) VALUES ($1, $2, $3, $4)`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) Delete(uid int64, scope string) error {
	query := `DELETE from tokens where scope=$1 and user_id=$2`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, scope, uid)
	return err
}
