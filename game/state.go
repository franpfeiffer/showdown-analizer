package game

type Move struct {
	Name  string
	Type  string
	Power int
}

type Pokemon struct {
	Name    string
	HP      int
	MaxHP   int
	Fainted bool
	Moves   []Move
	Status  string
	Ability string
	Boosts  map[string]int
	Type    []string
}

type Player struct {
	ID     string
	Name   string
	Team   map[string]*Pokemon
	Active *Pokemon
}

type BattleState struct {
	Players      map[string]*Player
	Turn         int
	Weather      string
	FieldEffects map[string]bool
}

func NewBattleState() *BattleState {
	return &BattleState{
		Players:      make(map[string]*Player),
		Turn:         0,
		Weather:      "",
		FieldEffects: make(map[string]bool),
	}
}
