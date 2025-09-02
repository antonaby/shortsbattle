package services

import (
	"os"
	"testing"
	"time"
)

func TestCreatingAndValidatingJwt(t *testing.T) {
	os.Setenv("JWK_KEY_SECRET", "eeGC9hxKwDnbI7ec2BluU16qSw3AporSP4ttxIFpfm4=")

	keyManager, err := NewKeyManager()
	if err != nil {
		t.Fatalf("key manager not created %v", err)
	}

	playerService := PlayersService{}

	// TODO: use interface instead
	authService := NewAuthService(keyManager, &playerService, AuthConfig{
		TokenExpTime: 5 * time.Minute,
		Issuer:       "shortsbattle",
		Audience:     "tg-mini-app",
	})

	rawToken, err := authService.NewTokenFromTgInitData("userId=1")
	if err != nil {
		t.Fatalf("failed to create JWT %v", err)
	}

	if len(rawToken) == 0 {
		t.Fatalf("token not created")
	}

	parsedToken, err := authService.ParseAndValidateJwt(rawToken)
	if err != nil {
		t.Fatalf("failed to parse JWT %v", err)
	}

	subject, ok := parsedToken.Subject()
	if !ok {
		t.Fatal("token has no issuer")
	}

	if  subject != "1" {
		t.Errorf("wring subject: %s", subject)
	}
}
