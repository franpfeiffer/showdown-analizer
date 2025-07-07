package client

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	showdownServerURL = "wss://sim3.psim.us/showdown/websocket"
)

type ShowdownClient struct {
	Conn *websocket.Conn
}

func NewShowdownClient() (*ShowdownClient, error) {
	u, err := url.Parse(showdownServerURL)
	if err != nil {
		return nil, fmt.Errorf("error al parsear la url del server: %w", err)
	}

	log.Printf("Conectando a %s", u.String())

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	c, resp, err := dialer.Dial(u.String(), nil)
	if err != nil {
		if resp != nil {
			log.Printf("HTTP Response Status: %s", resp.Status)
		}
		return nil, fmt.Errorf("error al conectar con el websocket: %w", err)
	}

	client := &ShowdownClient{Conn: c}
	log.Println("conectado exitosamente al servidor de showdown.")

	c.SetPingHandler(func(appData string) error {
		log.Printf("Received ping, sending pong")
		return c.WriteMessage(websocket.PongMessage, []byte(appData))
	})

	c.SetCloseHandler(func(code int, text string) error {
		log.Printf("Connection closed: %d %s", code, text)
		return nil
	})

	return client, nil
}

func (sc *ShowdownClient) Listen() {
	go func() {
		defer sc.Conn.Close()
		for {
			_, message, err := sc.Conn.ReadMessage()
			if err != nil {
				log.Println("error de lectura:", err)
				return
			}
			log.Printf("recibido: %s", message)
		}
	}()
}

func (sc *ShowdownClient) Send(message string) error {
	log.Printf("enviando: %s", message)
	return sc.Conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func (sc *ShowdownClient) JoinRoom(roomID string) error {
	return sc.Send(fmt.Sprintf("|/join %s", roomID))
}
