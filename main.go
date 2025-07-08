package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	log.Printf("Received connection request from %s", r.RemoteAddr)

	roomID := r.URL.Query().Get("roomid")
	log.Printf("Room ID requested: %s", roomID)

	if !strings.HasPrefix(roomID, "battle-") {
		roomID = "battle-" + roomID
		log.Printf("Room ID corregido a: %s", roomID)
	}

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

	messages := make(chan string, 100)
	done := make(chan struct{})

	defer func() {
		log.Printf("Cleaning up SSE connection for room %s", roomID)
		select {
		case <-done:
		default:
			close(done)
		}
		close(messages)
	}()

	battleState := game.NewBattleState()

	reconnectAttempts := 0
	const maxReconnects = 3
	var sdClient *client.ShowdownClient
	var err error

reconnect:
	log.Printf("Creating Showdown client (attempt %d)...", reconnectAttempts+1)
	sdClient, err = client.NewShowdownClient()
	if err != nil {
		log.Printf("Error al conectar con Showdown: %v", err)
		fmt.Fprintf(w, "data: <p>Error al conectar con Showdown: %v</p>\n\n", err)
		flusher.Flush()
		reconnectAttempts++
		if reconnectAttempts < maxReconnects {
			time.Sleep(2 * time.Second)
			goto reconnect
		}
		return
	}
	defer func() {
		if sdClient != nil && sdClient.Conn != nil {
			sdClient.Conn.Close()
		}
	}()

	log.Printf("Showdown client created successfully")

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

	go func() {
		defer func() {
			log.Printf("Closing Showdown message reader for room %s", roomID)
			select {
			case <-done:
			default:
				close(done)
			}
		}()

		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			default:
				_, message, err := sdClient.Conn.ReadMessage()
				if err != nil {
					log.Printf("Error reading message from Showdown: %v", err)
					select {
					case messages <- "__ERROR__:" + err.Error():
					case <-done:
					case <-ctx.Done():
					}
					return
				}
				log.Printf("Received message from Showdown: %s", string(message))
				select {
				case messages <- string(message):
				case <-done:
				case <-ctx.Done():
				}
			}
		}
	}()

	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()
	go func() {
		for {
			select {
			case <-pingTicker.C:
				select {
				case messages <- "__PING__":
				case <-done:
					return
				case <-ctx.Done():
					return
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled for room %s", roomID)
			return
		case msg, ok := <-messages:
			if !ok {
				log.Printf("Messages channel closed for room %s", roomID)
				return
			}
			if msg == "__PING__" {
				fmt.Fprintf(w, ": ping\n\n")
				flusher.Flush()
				continue
			}
			if strings.HasPrefix(msg, "__ERROR__:") {
				reconnectAttempts++
				if reconnectAttempts < maxReconnects {
					fmt.Fprintf(w, "data: <p>Reconectando con Showdown... (intento %d/%d)</p>\n\n", reconnectAttempts+1, maxReconnects)
					flusher.Flush()
					time.Sleep(2 * time.Second)
					goto reconnect
				} else {
					fmt.Fprintf(w, "data: <p class='error'>Error persistente al conectar con Showdown: %s</p>\n\n", msg[len("__ERROR__:"):])
					flusher.Flush()
					return
				}
			}
			lines := strings.Split(msg, "\n")
			var anyLogSent bool
			var battleEnded bool
			for _, line := range lines {
				if strings.HasPrefix(line, "|turn|") ||
					strings.HasPrefix(line, "|move|") ||
					strings.HasPrefix(line, "|switch|") ||
					strings.HasPrefix(line, "|damage|") ||
					strings.HasPrefix(line, "|faint|") ||
					strings.HasPrefix(line, "|start|") ||
					strings.HasPrefix(line, "|upkeep|") ||
					strings.HasPrefix(line, "|win|") ||
					strings.HasPrefix(line, "|lose|") ||
					strings.HasPrefix(line, "|player|") {
					parser.ProcessLine(battleState, line)
					log.Printf("Enviando al frontend: %s", line)
					fmt.Fprintf(w, "data: <p class='logline'>%s</p>\n\n", template.HTMLEscapeString(line))
					flusher.Flush()
					anyLogSent = true
					if strings.HasPrefix(line, "|win|") || strings.HasPrefix(line, "|lose|") {
						battleEnded = true
					}
				}
			}
			if anyLogSent {
				summary := parser.RenderBattleState(battleState)
				fmt.Fprintf(w, "data: %s\n\n", summary)
				flusher.Flush()
			}
			if battleEnded {
				log.Printf("Batalla terminada en room %s, cerrando SOLO esta conexión SSE.", roomID)
				fmt.Fprintf(w, "data: <p class='success'>¡Batalla terminada! El servidor sigue funcionando para nuevas conexiones.</p>\n\n")
				flusher.Flush()
				return
			}
		case <-done:
			log.Printf("Done signal received for room %s", roomID)
			return
		}
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

func main() {
	log.Printf("Iniciando servidor Showdown Analyzer...")

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

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := getPort()
	log.Printf("Servidor iniciado en puerto %s", port)

	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Servidor listo para recibir conexiones en puerto %s", port)
	log.Fatal(srv.ListenAndServe())
}
