package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type Captcha struct {
	ID        string
	Code      string
	ExpiresAt time.Time
}

type CaptchaStore struct {
	mu    sync.Mutex
	items map[string]Captcha
	ttl   time.Duration
}

func NewCaptchaStore(ttl time.Duration) *CaptchaStore {
	return &CaptchaStore{
		items: make(map[string]Captcha),
		ttl:   ttl,
	}
}

func (s *CaptchaStore) Generate() (Captcha, error) {
	code, err := generateDigitCode(6)
	if err != nil {
		return Captcha{}, err
	}

	id, err := generateCaptchaID()
	if err != nil {
		return Captcha{}, err
	}

	captcha := Captcha{ID: id, Code: code, ExpiresAt: time.Now().Add(s.ttl)}

	s.mu.Lock()
	s.items[id] = captcha
	s.mu.Unlock()

	return captcha, nil
}

func (s *CaptchaStore) Verify(id, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	captcha, ok := s.items[id]
	if !ok {
		return false
	}
	if time.Now().After(captcha.ExpiresAt) {
		delete(s.items, id)
		return false
	}
	if captcha.Code != code {
		return false
	}
	delete(s.items, id)
	return true
}

func generateDigitCode(length int) (string, error) {
	limit := big.NewInt(1000000)
	number, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", fmt.Errorf("generate captcha code: %w", err)
	}
	return fmt.Sprintf("%0*d", length, number.Int64()), nil
}

func generateCaptchaID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate captcha id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}