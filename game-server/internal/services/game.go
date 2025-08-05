package services

import "fmt"

type GameService struct {
}

func NewGameService() *GameService {
	return &GameService{}
}

func (g GameService) InitGames() error {
	fmt.Println("Games Initialized")
	return nil
}
