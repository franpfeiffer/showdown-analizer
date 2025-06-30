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

var pokemonDB map[string]PokemonData
var moveDB map[string]MoveData

func LoadPokemonData(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var pokes []PokemonData
	if err := json.NewDecoder(file).Decode(&pokes); err != nil {
		return err
	}
	pokemonDB = make(map[string]PokemonData)
	for _, p := range pokes {
		pokemonDB[strings.ToLower(p.Name)] = p
	}
	return nil
}

func LoadMoveData(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var moves []MoveData
	if err := json.NewDecoder(file).Decode(&moves); err != nil {
		return err
	}
	moveDB = make(map[string]MoveData)
	for _, m := range moves {
		moveDB[strings.ToLower(m.Name)] = m
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
