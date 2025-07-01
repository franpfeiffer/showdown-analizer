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
	log.Printf("Received connection request from %s", r.RemoteAddr)

	roomID := r.URL.Query().Get("roomid")
	log.Printf("Room ID requested: %s", roomID)

	if roomID == "" {
		log.Printf("Error: Empty room ID")
		http.Error(w, "El ID de la sala no puede estar vacío", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("Error: Streaming not supported")
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}

	messages := make(chan string)

	log.Printf("Creating Showdown client...")
	sdClient, err := client.NewShowdownClient()
	if err != nil {
		log.Printf("Error al conectar con Showdown: %v", err)
		fmt.Fprintf(w, "data: <p>Error al conectar con Showdown: %v</p>\n\n", err)
		flusher.Flush()
		return
	}
	defer sdClient.Conn.Close()

	log.Printf("Showdown client created successfully")
	battleState := game.NewBattleState()

	go func() {
		defer close(messages)
		log.Printf("Starting message reader goroutine")
		for {
			_, message, err := sdClient.Conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message from Showdown: %v", err)
				return
			}
			log.Printf("Received message from Showdown: %s", string(message))
			messages <- string(message)
		}
	}()

	log.Printf("Attempting to join room: %s", roomID)
	if err := sdClient.JoinRoom(roomID); err != nil {
		log.Printf("Error al unirse a la sala: %v", err)
		fmt.Fprintf(w, "data: <p>Error al unirse a la sala: %v</p>\n\n", err)
		flusher.Flush()
		return
	}

	log.Printf("Successfully joined room: %s", roomID)
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
	if err := data.LoadPokemonData("data/pokedex.json"); err != nil {
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

	fmt.Println("Servidor iniciado en http://localhost:8081")
	srv := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
