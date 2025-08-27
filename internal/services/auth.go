package services

import "github.com/antonaby/shortsbattle/game-server/internal/db"

type AuthService struct {
	txm db.TxManager
}

func NewAuthService(txm db.TxManager) *AuthService {
	return &AuthService{
		txm: txm,
	}
}

