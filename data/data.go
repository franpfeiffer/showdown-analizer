package data

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PokemonData struct {
	Name  string   `json:"name"`
	Types []string `json:"types"`
}

type MoveData struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Power int    `json:"power"`
}

type RawPokemonData struct {
	Name  string   `json:"name"`
	Types []string `json:"types"`
}

type RawMoveData struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Power int    `json:"basePower"`
}

var pokemonDB map[string]PokemonData
var moveDB map[string]MoveData

func LoadPokemonData(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var rawData map[string]RawPokemonData
	if err := json.NewDecoder(file).Decode(&rawData); err != nil {
		return err
	}

	pokemonDB = make(map[string]PokemonData)
	for _, p := range rawData {
		pokemonDB[strings.ToLower(p.Name)] = PokemonData{
			Name:  p.Name,
			Types: p.Types,
		}
	}
	return nil
}

func LoadMoveData(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var rawData map[string]RawMoveData
	if err := json.NewDecoder(file).Decode(&rawData); err != nil {
		return err
	}

	moveDB = make(map[string]MoveData)
	for _, m := range rawData {
		moveDB[strings.ToLower(m.Name)] = MoveData{
			Name:  m.Name,
			Type:  m.Type,
			Power: m.Power,
		}
	}
	return nil
}

func GetPokemonTypes(name string) []string {
	if p, ok := pokemonDB[strings.ToLower(name)]; ok {
		return p.Types
	}
	return nil
}

func GetMoveTypeAndPower(name string) (string, int, error) {
	if m, ok := moveDB[strings.ToLower(name)]; ok {
		return m.Type, m.Power, nil
	}
	return "", 80, fmt.Errorf("movimiento no encontrado: %s", name)
}

func GetAllMoves() []MoveData {
	var moves []MoveData
	for _, move := range moveDB {
		moves = append(moves, move)
	}
	return moves
}

func GetPokemonMovepool(pokemonName string) []MoveData {
	return GetAllMoves()
}
