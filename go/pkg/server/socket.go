package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Socket interface {
    GetOutBound() chan<- *Message
    GetInBound() <-chan *Message
    Close() error
}

type SocketImpl struct {
    outBound chan<- *Message
    inBound <-chan *Message
    conn *websocket.Conn
}

func (s *SocketImpl) GetOutBound() chan<- *Message {
    return s.outBound
}

func (s *SocketImpl) GetInBound() <-chan *Message {
    return s.inBound
}

func (s *SocketImpl) Close() error {
    return s.conn.Close()
}

func NewSocket(w http.ResponseWriter, r *http.Request) (Socket, error) {
	c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return nil, err;
    }

    // from me to network
    out := make(chan *Message)

    // from network to me
    in := make(chan *Message) // other type

    go func() {
        defer func() {
            // TODO: Do we need to create a close message?
            // in <- CloseMessage()
            c.Close()
        }()

        for {
            mt, message, err := c.ReadMessage()
            if err != nil {
                log.Println("read:", err)
                break
            }

            if mt != websocket.TextMessage {
                continue
            }

            in <- FromSocket(message)
        }
    }()

    go func() {
        for msg := range out {
            msg, err := json.Marshal(msg.Message)
            if err != nil {
                log.Fatalf("%+v\n", err)
            }

            c.WriteMessage(websocket.TextMessage, msg)
        }
    }()

    return &SocketImpl{
        out,
        in,
        c,
    }, nil
}

