package parser

import (
	"showdown-analizer/data"
	"showdown-analizer/game"
	"strconv"
	"strings"
)

func ParseLog(logText string) (*game.BattleState, error) {
	state := game.NewBattleState()
	lines := strings.Split(logText, "\n")

	for _, line := range lines {
		ProcessLine(state, line)
	}

	return state, nil
}

func ProcessLine(state *game.BattleState, line string) {
	parts := strings.Split(strings.TrimSpace(line), "|")
	if len(parts) < 2 {
		return
	}
	switch parts[1] {
	case "player":
		if len(parts) >= 4 {
			id := parts[2]
			name := parts[3]
			if _, ok := state.Players[id]; !ok {
				state.Players[id] = &game.Player{
					ID:   id,
					Name: name,
					Team: make(map[string]*game.Pokemon),
				}
			}
		}
	case "poke":
		if len(parts) >= 4 {
			id := parts[2] // p1
			pokeInfo := strings.Split(parts[3], ",")
			name := strings.TrimSpace(pokeInfo[0])
			types := data.GetPokemonTypes(name)
			if player, ok := state.Players[id]; ok {
				player.Team[name] = &game.Pokemon{Name: name, Type: types}
			}
		}
	case "team":
		if len(parts) >= 5 {
			id := parts[2]
			pokeName := parts[3]
			moveNames := strings.Split(parts[4], ", ")
			moves := []game.Move{}
			for _, mn := range moveNames {
				type_, power, _ := data.GetMoveTypeAndPower(mn)
				moves = append(moves, game.Move{Name: mn, Type: type_, Power: power})
			}
			if player, ok := state.Players[id]; ok {
				if poke, ok := player.Team[pokeName]; ok {
					poke.Moves = moves
				}
			}
		}
	case "switch":
		if len(parts) >= 4 {
			activeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(activeInfo) == 2 {
				playerID := string(activeInfo[0][:2])
				pokeName := activeInfo[1]
				types := data.GetPokemonTypes(pokeName)
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						poke.Type = types
						player.Active = poke
						if len(parts) >= 4 {
							hpInfo := strings.Split(parts[4-1], "/")
							if len(hpInfo) == 2 {
								hp, _ := strconv.Atoi(strings.TrimSpace(hpInfo[0]))
								maxhp, _ := strconv.Atoi(strings.TrimSpace(hpInfo[1]))
								poke.HP = hp
								poke.MaxHP = maxhp
							}
						}
					}
				}
			}
		}
	case "move":
		if len(parts) >= 4 {
			userInfo := strings.SplitN(parts[2], ": ", 2)
			if len(userInfo) == 2 {
				playerID := string(userInfo[0][:2])
				moveName := parts[3]
				type_, power, _ := data.GetMoveTypeAndPower(moveName)
				move := game.Move{Name: moveName, Type: type_, Power: power}
				if player, ok := state.Players[playerID]; ok && player.Active != nil {
					exists := false
					for _, m := range player.Active.Moves {
						if m.Name == move.Name {
							exists = true
							break
						}
					}
					if !exists {
						player.Active.Moves = append(player.Active.Moves, move)
					}
				}
			}
		}
	case "damage":
		if len(parts) >= 4 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						hpInfo := strings.Split(parts[3], "/")
						if len(hpInfo) == 2 {
							hp, _ := strconv.Atoi(strings.TrimSpace(hpInfo[0]))
							maxhp, _ := strconv.Atoi(strings.TrimSpace(hpInfo[1]))
							poke.HP = hp
							poke.MaxHP = maxhp
						}
					}
				}
			}
		}
	case "faint":
		if len(parts) >= 3 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						poke.Fainted = true
					}
				}
			}
		}
	case "turn":
		if len(parts) >= 3 {
			t, err := strconv.Atoi(parts[2])
			if err == nil {
				state.Turn = t
			}
		}
	case "-status":
		if len(parts) >= 4 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				status := parts[3]
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						poke.Status = status
					}
				}
			}
		}
	case "-curestatus":
		if len(parts) >= 4 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						poke.Status = ""
					}
				}
			}
		}
	case "-boost":
		if len(parts) >= 5 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				stat := parts[3]
				amount, _ := strconv.Atoi(parts[4])
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						if poke.Boosts == nil {
							poke.Boosts = make(map[string]int)
						}
						poke.Boosts[stat] += amount
					}
				}
			}
		}
	case "-unboost":
		if len(parts) >= 5 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				stat := parts[3]
				amount, _ := strconv.Atoi(parts[4])
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						if poke.Boosts == nil {
							poke.Boosts = make(map[string]int)
						}
						poke.Boosts[stat] -= amount
					}
				}
			}
		}
	case "-setboost":
		if len(parts) >= 5 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				stat := parts[3]
				amount, _ := strconv.Atoi(parts[4])
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						if poke.Boosts == nil {
							poke.Boosts = make(map[string]int)
						}
						poke.Boosts[stat] = amount
					}
				}
			}
		}
	case "-weather":
		if len(parts) >= 3 {
			state.Weather = parts[2]
		}
	case "-fieldstart":
		if len(parts) >= 3 {
			effect := parts[2]
			state.FieldEffects[effect] = true
		}
	case "-fieldend":
		if len(parts) >= 3 {
			effect := parts[2]
			delete(state.FieldEffects, effect)
		}
	case "-ability":
		if len(parts) >= 4 {
			pokeInfo := strings.SplitN(parts[2], ": ", 2)
			if len(pokeInfo) == 2 {
				playerID := string(pokeInfo[0][:2])
				pokeName := pokeInfo[1]
				ability := parts[3]
				if player, ok := state.Players[playerID]; ok {
					if poke, ok := player.Team[pokeName]; ok {
						poke.Ability = ability
					}
				}
			}
		}
	}
}
