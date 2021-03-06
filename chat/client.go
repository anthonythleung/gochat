package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/sirupsen/logrus"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client ... middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

// Message ... a message sent or received by a client
type Message struct {
	Type     string `json:"type"`
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log.WithFields(logrus.Fields{
					"error": err,
					"hubID": c.hub.id,
				}).Error("Websocket Closed Unexpectedly")
			}
			break
		}
		var parsedMessage Message
		err = json.Unmarshal(message, &parsedMessage)
		UUID, _ := uuid.NewV4()
		messageID := UUID.String()

		c.hub.messages <- message

		// Add message to cassandra log
		err = c.hub.session.Query(`INSERT INTO messages (channel_id, message_id, created_at, author_id, content, type) VALUES (?, ?, ?, ?, ?, ?)`,
			c.hub.id, messageID, time.Now(), parsedMessage.ID, parsedMessage.Message, parsedMessage.Type).Exec()
		if err != nil {
			c.hub.log.WithFields(logrus.Fields{
				"error":     err,
				"messageID": messageID,
				"hubID":     c.hub.id,
			}).Error("Error inserting message into cassandra")
		}

		// Add message to elastic search
		result, err := c.hub.elastic.Index().
			Index("messages").
			Type("message").
			Id(messageID).
			BodyJson(parsedMessage).
			Refresh("wait_for").
			Do(context.Background())

		if err != nil {
			c.hub.log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Error inserting message to elasticsearch")
		} else {

			c.hub.log.WithFields(logrus.Fields{
				"resultID":    result.Id,
				"resultIndex": result.Index,
				"resultType":  result.Type,
			}).Info("Indexed message to elasticsearch")
		}

	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(newhub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: newhub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
