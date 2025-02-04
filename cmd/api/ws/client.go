package ws

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/cmd/api/ws/channels"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/gorilla/websocket"
)

// ClientHandler -
type ClientHandler func(*Client, []byte) error

// ClientEvent -
type ClientEvent func([]byte) error

// Client - nolint
type Client struct {
	id   int
	conn *websocket.Conn

	sender chan channels.Message
	stop   chan struct{}

	subscriptions sync.Map

	handlers map[string]ClientHandler
	sendMux  sync.Mutex

	onSubscribe   ClientEvent //nolint
	onUnsubscribe ClientEvent //nolint

	hub *Hub
}

// NewClient -
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		id:   rand.Int(),
		conn: conn,

		sender: make(chan channels.Message),
		stop:   make(chan struct{}),

		handlers: make(map[string]ClientHandler),
	}
}

// GetSubscription -
func (c *Client) GetSubscription(name string) (channels.Channel, bool) {
	val, ok := c.subscriptions.Load(name)
	if !ok {
		return nil, ok
	}
	channel, ok := val.(channels.Channel)
	return channel, ok
}

// Send -
func (c *Client) Send(msg channels.Message) {
	if _, ok := c.subscriptions.Load(msg.ChannelName); ok {
		c.sender <- msg
	}
}

// Run -
func (c *Client) Run() {
	go c.send()
	go c.receive()
}

// Close -
func (c *Client) Close() {
	close(c.stop)
	close(c.sender)
	c.conn.Close()
	c.hub.RemoveClient(c)
}

// AddHandler -
func (c *Client) AddHandler(name string, handler ClientHandler) {
	c.handlers[name] = handler
}

func (c *Client) sendMessage(message interface{}) error {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(time.Second * 2)); err != nil {
		return err
	}
	return c.conn.WriteJSON(message)
}

func (c *Client) sendError(err error) {
	msg := StatusMessage{
		Status: ErrorStatus,
		Text:   err.Error(),
	}
	if err := c.sendMessage(msg); err != nil {
		logger.Error(err)
	}
}

func (c *Client) sendOk(text string) error {
	msg := StatusMessage{
		Status: OkStatus,
		Text:   text,
	}
	return c.sendMessage(msg)
}

func (c *Client) receive() {
	for {
		select {
		case <-c.stop:
			return
		default:
			if err := c.conn.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
				logger.Error(err)
				continue
			}
			messageType, data, err := c.conn.ReadMessage()
			if err != nil {
				c.Close()
				continue
			}

			if messageType != websocket.TextMessage {
				continue
			}

			var msg ActionMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				logger.Error(err)
				continue
			}

			handler, ok := c.handlers[msg.Action]
			if !ok {
				c.sendError(fmt.Errorf("Unknown handler action: %s", msg.Action))
				continue
			}
			if err := handler(c, data); err != nil {
				c.sendError(err)
				continue
			}
		}
	}
}

func (c *Client) send() {
	for {
		select {
		case <-c.stop:
			return
		case msg := <-c.sender:
			if msg.ChannelName != "" && msg.Body != nil {
				if err := c.sendMessage(msg); err != nil {
					c.sendError(err)
				}
			}
		}
	}
}
