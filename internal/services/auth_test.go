package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/tests"
	"github.com/stretchr/testify/mock"
)

func TestCreateAndValidateJwt(t *testing.T) {
	os.Setenv("JWK_KEY_SECRET", "eeGC9hxKwDnbI7ec2BluU16qSw3AporSP4ttxIFpfm4=")
	ctx := context.Background()

	keyManager, err := NewKeyManager()
	if err != nil {
		t.Fatalf("key manager not created %v", err)
	}

	mockPlayerService := new(tests.MockPlayerService)
	mockPlayerService.
		On("CheckPlayerExistsOrCreate", mock.Anything, mock.Anything).
		Return(&qg.Player{TgID: 1}, nil)

	authService := NewAuthService(keyManager, mockPlayerService, AuthConfig{
		TokenExpTime: 5 * time.Minute,
		Issuer:       "shortsbattle",
		Audience:     "tg-mini-app",
	})

	rawToken, err := authService.NewTokenFromTgInitData(ctx, "userId=1")
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

	if subject != "1" {
		t.Errorf("wring subject: %s", subject)
	}
}
