package services


type GameInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GameService struct {

}

func NewGameService() *GameService {
	return &GameService{}
}

func (g GameService) GetGames() []GameInfo {
	return []GameInfo{
		{
			Name:        "The most cute cat 🐈",
			Description: "A game about the cutest cat in the world. 🐱",
		},
		{
			Name:        "Funniest fail video 😂",
			Description: "Submit a hilarious fail that makes everyone laugh!",
		},
		{
			Name:        "Best dance move 💃",
			Description: "Show off your craziest or smoothest dance step.",
		},
		{
			Name:        "Unexpected twist 🎭",
			Description: "Videos that take a surprising turn. Shock us!",
		},
		{
			Name:        "Cutest baby animal 🐾",
			Description: "Puppies, kittens, ducklings... bring the awws!",
		},
		{
			Name:        "Most epic moment ⚡",
			Description: "Highlight something legendary, heroic, or just cool.",
		},
		{
			Name:        "Mind-blowing magic trick 🎩✨",
			Description: "Is it real? Is it edited? Blow our minds!",
		},
		{
			Name:        "Satisfying video 🍰",
			Description: "Soap cutting, symmetry, pouring — we want chill.",
		},
		{
			Name:        "Cringe overload 😬",
			Description: "Bring the secondhand embarrassment in a fun way.",
		},
		{
			Name:        "Best pet reaction 🐶😲",
			Description: "Pets doing something wild, unexpected, or smart!",
		},
	}
}
