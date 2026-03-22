package uplink

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/minicodemonkey/chief/internal/protocol"
)

// Client is a WebSocket client that connects to the Uplink server.
type Client struct {
	url   string
	token string

	conn      *websocket.Conn
	mu        sync.Mutex
	sessionID string
	ready     chan struct{}
	readyOnce sync.Once
	onMessage func(protocol.Envelope)
	done      chan struct{}
	closed    bool
}

// NewClient creates a new WebSocket client for the given URL and auth token.
func NewClient(url, token string) *Client {
	return &Client{
		url:   url,
		token: token,
		ready: make(chan struct{}),
		done:  make(chan struct{}),
	}
}

// OnMessage sets the callback for incoming envelopes.
// Must be called before Connect.
func (c *Client) OnMessage(fn func(protocol.Envelope)) {
	c.onMessage = fn
}

// Ready returns a channel that closes when the welcome message has been received.
func (c *Client) Ready() <-chan struct{} {
	return c.ready
}

// SessionID returns the connection/session ID from the welcome message.
func (c *Client) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

// Connect dials the WebSocket server with the Bearer token and starts reading messages.
func (c *Client) Connect() error {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)

	conn, _, err := websocket.DefaultDialer.Dial(c.url, header)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	go c.readLoop()

	return nil
}

// Send marshals an envelope and writes it to the WebSocket connection.
func (c *Client) Send(env protocol.Envelope) error {
	data, err := env.Marshal()
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("websocket write: %w", err)
	}

	return nil
}

// Close gracefully closes the WebSocket connection and stops reconnection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true

	if c.conn == nil {
		return nil
	}

	err := c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		// Connection may already be closed; ignore write errors and close.
		_ = c.conn.Close()
		c.conn = nil
		return nil
	}

	err = c.conn.Close()
	c.conn = nil
	return err
}

func (c *Client) readLoop() {
	defer close(c.done)

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		env, err := protocol.ParseEnvelope(data)
		if err != nil {
			continue
		}

		if env.Type == protocol.TypeWelcome {
			welcome, err := protocol.DecodePayload[protocol.Welcome](env)
			if err == nil {
				c.mu.Lock()
				c.sessionID = welcome.ConnectionID
				c.mu.Unlock()
				c.readyOnce.Do(func() { close(c.ready) })
			}
		}

		if c.onMessage != nil {
			c.onMessage(env)
		}
	}
}

// ConnectWithReconnect connects and auto-reconnects with exponential backoff
// (jitter, capped at 60s). The onConnect callback is called after each successful
// welcome handshake.
func (c *Client) ConnectWithReconnect(onConnect func()) error {
	return c.connectWithReconnect(onConnect, nil)
}

// connectWithReconnect is the internal implementation that accepts a sleep function for testing.
func (c *Client) connectWithReconnect(onConnect func(), sleepFn func(time.Duration)) error {
	if sleepFn == nil {
		sleepFn = time.Sleep
	}

	attempt := 0
	for {
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return nil
		}
		// Reset ready channel for this connection attempt.
		c.ready = make(chan struct{})
		c.readyOnce = sync.Once{}
		c.done = make(chan struct{})
		c.mu.Unlock()

		err := c.Connect()
		if err != nil {
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return nil
			}
			c.mu.Unlock()
			attempt++
			sleepFn(backoff(attempt))
			continue
		}

		// Wait for welcome.
		select {
		case <-c.Ready():
			attempt = 0
			if onConnect != nil {
				onConnect()
			}
		case <-c.done:
			// Connection closed before welcome; reconnect.
			attempt++
			sleepFn(backoff(attempt))
			continue
		}

		// Wait for connection to drop.
		<-c.done

		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return nil
		}
		c.mu.Unlock()

		// Connection lost; reconnect.
		attempt++
		sleepFn(backoff(attempt))
	}
}

// backoff returns a duration with exponential backoff and jitter, capped at 60s.
func backoff(attempt int) time.Duration {
	base := math.Pow(2, float64(attempt))
	if base > 60 {
		base = 60
	}
	jitter := rand.Float64() * base //nolint:gosec
	return time.Duration(jitter * float64(time.Second))
}
