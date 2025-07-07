package parser

import (
	"fmt"
	"log"
	"showdown-analizer/game"
	"sort"
	"strings"
	"unicode"
)

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	return string(runes)
}

var typeChart = map[string]map[string]float64{
	"Fire": {
		"Water": 0.5, "Rock": 0.5, "Fire": 0.5, "Grass": 2, "Ice": 2, "Bug": 2, "Steel": 2, "Dragon": 0.5,
	},
	"Flying": {
		"Grass": 2, "Fighting": 2, "Bug": 2, "Electric": 0.5, "Rock": 0.5, "Steel": 0.5,
	},
	"Dragon": {
		"Dragon": 2, "Steel": 0.5,
	},
	"Water": {
		"Fire": 2, "Water": 0.5, "Grass": 0.5, "Ground": 2, "Rock": 2, "Dragon": 0.5,
	},
	"Dark": {
		"Ghost": 2, "Psychic": 2, "Dark": 0.5, "Fighting": 0.5, "Fairy": 0.5,
	},
	"Rock": {
		"Fire": 2, "Ice": 2, "Flying": 2, "Bug": 2, "Fighting": 0.5, "Ground": 0.5, "Steel": 0.5,
	},
	"Ice": {
		"Dragon": 2, "Flying": 2, "Grass": 2, "Ground": 2, "Fire": 0.5, "Water": 0.5, "Ice": 0.5, "Steel": 0.5,
	},
	"Steel": {
		"Rock": 2, "Ice": 2, "Fairy": 2, "Steel": 0.5, "Fire": 0.5, "Water": 0.5, "Electric": 0.5,
	},
	"Fighting": {
		"Normal": 2, "Rock": 2, "Steel": 2, "Ice": 2, "Dark": 2, "Ghost": 0, "Poison": 0.5, "Flying": 0.5, "Psychic": 0.5, "Bug": 0.5, "Fairy": 0.5,
	},
	"Normal": {
		"Rock": 0.5, "Ghost": 0, "Steel": 0.5,
	},
	"Electric": {
		"Flying": 2, "Water": 2, "Ground": 0, "Grass": 0.5, "Electric": 0.5, "Dragon": 0.5,
	},
	"Grass": {
		"Ground": 2, "Rock": 2, "Water": 2, "Flying": 0.5, "Poison": 0.5, "Bug": 0.5, "Steel": 0.5, "Fire": 0.5, "Grass": 0.5, "Dragon": 0.5,
	},
	"Psychic": {
		"Fighting": 2, "Poison": 2, "Steel": 0.5, "Psychic": 0.5, "Dark": 0,
	},
	"Ghost": {
		"Ghost": 2, "Psychic": 2, "Normal": 0, "Dark": 0.5,
	},
	"Poison": {
		"Grass": 2, "Fairy": 2, "Poison": 0.5, "Ground": 0.5, "Rock": 0.5, "Ghost": 0.5, "Steel": 0,
	},
	"Ground": {
		"Poison": 2, "Rock": 2, "Steel": 2, "Fire": 2, "Electric": 2, "Flying": 0, "Bug": 0.5, "Grass": 0.5,
	},
	"Bug": {
		"Grass": 2, "Psychic": 2, "Dark": 2, "Fighting": 0.5, "Flying": 0.5, "Poison": 0.5, "Ghost": 0.5, "Steel": 0.5, "Fire": 0.5, "Fairy": 0.5,
	},
	"Fairy": {
		"Fighting": 2, "Dragon": 2, "Dark": 2, "Poison": 0.5, "Steel": 0.5, "Fire": 0.5,
	},
}

func getTypeEffectiveness(moveType string, targetTypes []string) float64 {
	eff := 1.0
	for _, t := range targetTypes {
		if m, ok := typeChart[moveType]; ok {
			if v, ok := m[t]; ok {
				eff *= v
			}
		}
	}
	return eff
}

func getWeaknesses(pokemonTypes []string) []string {
	weaknesses := make(map[string]bool)
	
	for attackType, effectiveness := range typeChart {
		for _, defenseType := range pokemonTypes {
			if eff, exists := effectiveness[defenseType]; exists && eff > 1 {
				weaknesses[attackType] = true
			}
		}
	}
	
	var result []string
	for weakness := range weaknesses {
		result = append(result, weakness)
	}
	sort.Strings(result)
	return result
}

func getSuggestions(player *game.Player, opponent *game.Pokemon) string {
	if player.Active == nil || len(player.Active.Moves) == 0 {
		return "<i>Sin movimientos conocidos aún.</i>"
	}
	
	type moveScore struct {
		move  game.Move
		score float64
		eff   float64
	}
	
	var scored []moveScore
	for _, move := range player.Active.Moves {
		power := move.Power
		if power == 0 {
			power = 80
		}
		eff := getTypeEffectiveness(move.Type, opponent.Type)
		score := float64(power) * eff
		scored = append(scored, moveScore{move, score, eff})
	}
	
	sort.Slice(scored, func(i, j int) bool { 
		return scored[i].score > scored[j].score 
	})
	
	var result strings.Builder
	result.WriteString("Movimientos conocidos:<br>")
	for i, ms := range scored {
		effText := ""
		if ms.eff > 1 {
			effText = " (¡Súper efectivo!)"
		} else if ms.eff < 1 {
			effText = " (No muy efectivo)"
		}
		
		result.WriteString(fmt.Sprintf("%d. <b>%s</b> [%s] - %.0f pts%s<br>", 
			i+1, ms.move.Name, ms.move.Type, ms.score, effText))
	}
	
	return result.String()
}

func bestMove(p1 *game.Pokemon, p2 *game.Pokemon) (game.Move, float64) {
	best := game.Move{}
	bestScore := -1.0
	for _, move := range p1.Moves {
		power := move.Power
		if power == 0 {
			power = 80
		}
		eff := getTypeEffectiveness(move.Type, p2.Type)
		score := float64(power) * eff
		if score > bestScore {
			best = move
			bestScore = score
		}
	}
	return best, bestScore
}

func bestSwitch(p1 *game.Player, p2 *game.Pokemon) *game.Pokemon {
	var best *game.Pokemon
	bestScore := 0.0
	for _, poke := range p1.Team {
		if poke == p1.Active || poke.Fainted {
			continue
		}
		score := 1.0
		for _, t := range p2.Type {
			for _, myType := range poke.Type {
				if m, ok := typeChart[t]; ok {
					if v, ok := m[myType]; ok {
						score *= v
					}
				}
			}
		}
		if best == nil || score < bestScore {
			best = poke
			bestScore = score
		}
	}
	if bestScore < 1.0 {
		return best
	}
	return nil
}

func bestMovesList(p2 *game.Pokemon, p1 *game.Pokemon) []game.Move {
	type moveScore struct {
		move  game.Move
		score float64
	}
	var scored []moveScore
	for _, move := range p2.Moves {
		power := move.Power
		if power == 0 {
			power = 80
		}
		eff := getTypeEffectiveness(move.Type, p1.Type)
		score := float64(power) * eff
		scored = append(scored, moveScore{move, score})
	}
	sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })
	res := []game.Move{}
	for i := 0; i < len(scored) && i < 5; i++ {
		res = append(res, scored[i].move)
	}
	return res
}

func RenderBattleState(state *game.BattleState) string {
	var sb strings.Builder

	sb.WriteString("<div class='battle-summary'>")

	if state.Weather != "" {
		sb.WriteString(fmt.Sprintf("<div><b>Clima:</b> %s</div>", state.Weather))
	}
	if len(state.FieldEffects) > 0 {
		effects := make([]string, 0, len(state.FieldEffects))
		for eff := range state.FieldEffects {
			effects = append(effects, eff)
		}
		sort.Strings(effects)
		sb.WriteString("<div><b>Campo:</b> " + strings.Join(effects, ", ") + "</div>")
	}

	sb.WriteString(fmt.Sprintf("<h3>Turno: %d</h3>", state.Turn))

	p1 := state.Players["p1"]
	p2 := state.Players["p2"]

	for _, player := range state.Players {
		sb.WriteString(fmt.Sprintf("<h4>%s</h4>", player.Name))
		if player.Active != nil {
			poke := player.Active
			ps := "?/?"
			if poke.MaxHP > 0 {
				ps = fmt.Sprintf("%d/%d", poke.HP, poke.MaxHP)
			}
			fainted := ""
			if poke.Fainted {
				fainted = "<span style='color:#e74c3c;'>(Debilitado)</span>"
			}
			status := ""
			if poke.Status != "" {
				status = fmt.Sprintf("<span style='color:#f1c40f;'>[%s]</span>", poke.Status)
			}
			ability := ""
			if poke.Ability != "" {
				ability = fmt.Sprintf("<span style='color:#7ed6df;'>%s</span>", poke.Ability)
			}
			
			typeStr := ""
			if len(poke.Type) > 0 {
				typeStr = fmt.Sprintf(" <span style='color:#9b9b9b;'>(%s)</span>", strings.Join(poke.Type, "/"))
			}
			
			sb.WriteString(fmt.Sprintf("<b>%s</b>%s %s %s <span style='color:#aaa;'>[%s]</span> %s<br>", poke.Name, typeStr, fainted, status, ps, ability))
			
			if len(poke.Type) > 0 {
				weaknesses := getWeaknesses(poke.Type)
				if len(weaknesses) > 0 {
					sb.WriteString(fmt.Sprintf("<span style='color:#ff6b6b;'>Débil a: %s</span><br>", strings.Join(weaknesses, ", ")))
				}
			}
			
			if len(poke.Boosts) > 0 {
				boosts := make([]string, 0, len(poke.Boosts))
				for stat, val := range poke.Boosts {
					if val != 0 {
						prefix := "+"
						if val < 0 {
							prefix = ""
						}
						boosts = append(boosts, fmt.Sprintf("%s%d %s", prefix, val, capitalizeFirst(stat)))
					}
				}
				if len(boosts) > 0 {
					sb.WriteString("<span style='color:#e67e22;'>Boosts: " + strings.Join(boosts, ", ") + "</span><br>")
				}
			}
			if len(poke.Moves) > 0 {
				sb.WriteString("Movimientos vistos: ")
				moveNames := []string{}
				for _, m := range poke.Moves {
					moveNames = append(moveNames, m.Name)
				}
				sb.WriteString(strings.Join(moveNames, ", "))
				sb.WriteString("<br>")
			}
		}
	}

	if p1 != nil && p1.Active != nil {
		log.Printf("[Render] Movimientos conocidos de %s: %v", p1.Active.Name, p1.Active.Moves)
	}

	if p1 != nil && p2 != nil && p1.Active != nil && p2.Active != nil {
		sb.WriteString("<div class='suggestion'><b>Sugerencias para " + p1.Name + ":</b><br>")
		sb.WriteString(getSuggestions(p1, p2.Active))
		sb.WriteString("</div>")
		
		sb.WriteString("<div class='suggestion'><b>Sugerencias para " + p2.Name + ":</b><br>")
		sb.WriteString(getSuggestions(p2, p1.Active))
		sb.WriteString("</div>")
	}

	sb.WriteString("</div>")
	return sb.String()
}
