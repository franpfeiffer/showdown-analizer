package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"showdown-analizer/client"
	"showdown-analizer/data"
	"showdown-analizer/game"
	"showdown-analizer/parser"
	"strings"
	"time"
)

func parseTemplates() *template.Template {
	return template.Must(template.ParseGlob("templates/*.html"))
}

var templates = parseTemplates()

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "Error al renderizar la plantilla", http.StatusInternalServerError)
	}
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomid")
	if roomID == "" {
		http.Error(w, "El ID de la sala no puede estar vacío", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}

	messages := make(chan string)

	sdClient, err := client.NewShowdownClient()
	if err != nil {
		log.Printf("Error al conectar con Showdown: %v", err)
		fmt.Fprintf(w, "data: <p>Error al conectar con Showdown: %v</p>\n\n", err)
		flusher.Flush()
		return
	}
	defer sdClient.Conn.Close()
	battleState := game.NewBattleState()

	go func() {
		defer close(messages)
		for {
			_, message, err := sdClient.Conn.ReadMessage()
			if err != nil {
				log.Println("Desconectado del servidor de Showdown.")
				return
			}
			messages <- string(message)
		}
	}()

	if err := sdClient.JoinRoom(roomID); err != nil {
		log.Printf("Error al unirse a la sala: %v", err)
		fmt.Fprintf(w, "data: <p>Error al unirse a la sala: %v</p>\n\n", err)
		flusher.Flush()
		return
	}

	fmt.Fprintf(w, "data: <p>Conectado a la sala <strong>%s</strong>. Esperando eventos...</p>\n\n", roomID)
	flusher.Flush()

	for msg := range messages {
		lines := strings.Split(msg, "\n")
		for _, line := range lines {
			if len(line) > 1 {
				parser.ProcessLine(battleState, line)
				isRelevant := false
				if strings.HasPrefix(line, "|turn|") || strings.HasPrefix(line, "|move|") || strings.HasPrefix(line, "|switch|") || strings.HasPrefix(line, "|damage|") || strings.HasPrefix(line, "|faint|") {
					isRelevant = true
				}
				fmt.Fprintf(w, "data: <p class='logline'>%s</p>\n\n", template.HTMLEscapeString(line))
				flusher.Flush()
				if isRelevant {
					summary := parser.RenderBattleState(battleState)
					fmt.Fprintf(w, "data: %s\n\n", summary)
					flusher.Flush()
				}
			}
		}
	}

	log.Println("El cliente se ha desconectado.")
}

func main() {
	if err := data.LoadPokemonData("data/pokemon.json"); err != nil {
		log.Fatalf("Error cargando datos de Pokémon: %v", err)
	}
	if err := data.LoadMoveData("data/moves.json"); err != nil {
		log.Fatalf("Error cargando datos de movimientos: %v", err)
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/connect", handleConnect)

	fmt.Println("Servidor iniciado en http://localhost:8080")
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
