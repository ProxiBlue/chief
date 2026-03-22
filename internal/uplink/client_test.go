package uplink

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/minicodemonkey/chief/internal/protocol"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// sendEnvelope writes a protocol envelope as JSON over the WebSocket connection.
func sendEnvelope(t *testing.T, conn *websocket.Conn, env protocol.Envelope) {
	t.Helper()
	data, err := env.Marshal()
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write envelope: %v", err)
	}
}

// makeWelcome builds a welcome envelope with the given connection ID.
func makeWelcome(connectionID string) protocol.Envelope {
	payload, _ := json.Marshal(protocol.Welcome{
		ServerVersion: "1.0.0",
		ConnectionID:  connectionID,
	})
	env := protocol.NewEnvelope(protocol.TypeWelcome, "server")
	env.Payload = payload
	return env
}

func TestConnect_SendsAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		sendEnvelope(t, conn, makeWelcome("sess-1"))
		// Keep connection open until client disconnects.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client := NewClient(wsURL, "test-token-123")
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer client.Close()

	select {
	case <-client.Ready():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ready")
	}

	if gotAuth != "Bearer test-token-123" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer test-token-123")
	}
}

func TestConnect_WelcomeStoresSessionID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		sendEnvelope(t, conn, makeWelcome("session-abc"))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client := NewClient(wsURL, "tok")
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer client.Close()

	select {
	case <-client.Ready():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ready")
	}

	if got := client.SessionID(); got != "session-abc" {
		t.Errorf("SessionID() = %q, want %q", got, "session-abc")
	}
}

func TestSendAndReceive(t *testing.T) {
	var received []protocol.Envelope
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		sendEnvelope(t, conn, makeWelcome("sess-send"))

		// Echo back any messages received from the client.
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client := NewClient(wsURL, "tok")
	client.OnMessage(func(env protocol.Envelope) {
		mu.Lock()
		received = append(received, env)
		mu.Unlock()
	})

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer client.Close()

	select {
	case <-client.Ready():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ready")
	}

	// Send an envelope.
	env := protocol.NewEnvelope(protocol.TypeDeviceHeartbeat, "dev-1")
	if err := client.Send(env); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	// Wait for the echo response.
	deadline := time.After(2 * time.Second)
	for {
		mu.Lock()
		// Welcome + echoed heartbeat = at least 2 messages.
		count := len(received)
		mu.Unlock()
		if count >= 2 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for echo; got %d messages", count)
		case <-time.After(10 * time.Millisecond):
		}
	}

	mu.Lock()
	defer mu.Unlock()

	// First message is welcome, second is the echoed heartbeat.
	if received[0].Type != protocol.TypeWelcome {
		t.Errorf("first message type = %q, want %q", received[0].Type, protocol.TypeWelcome)
	}
	if received[1].Type != protocol.TypeDeviceHeartbeat {
		t.Errorf("echoed message type = %q, want %q", received[1].Type, protocol.TypeDeviceHeartbeat)
	}
	if received[1].DeviceID != "dev-1" {
		t.Errorf("echoed DeviceID = %q, want %q", received[1].DeviceID, "dev-1")
	}
}

func TestOnMessageCallback(t *testing.T) {
	callbackCh := make(chan protocol.Envelope, 10)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		sendEnvelope(t, conn, makeWelcome("sess-cb"))

		// Send an extra message after welcome.
		extra := protocol.NewEnvelope(protocol.TypeSync, "server")
		sendEnvelope(t, conn, extra)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client := NewClient(wsURL, "tok")
	client.OnMessage(func(env protocol.Envelope) {
		callbackCh <- env
	})

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer client.Close()

	// Collect two messages: welcome + sync.
	var msgs []protocol.Envelope
	timeout := time.After(2 * time.Second)
	for len(msgs) < 2 {
		select {
		case env := <-callbackCh:
			msgs = append(msgs, env)
		case <-timeout:
			t.Fatalf("timed out; got %d messages", len(msgs))
		}
	}

	if msgs[0].Type != protocol.TypeWelcome {
		t.Errorf("msgs[0].Type = %q, want welcome", msgs[0].Type)
	}
	if msgs[1].Type != protocol.TypeSync {
		t.Errorf("msgs[1].Type = %q, want sync", msgs[1].Type)
	}
}

func TestConnectWithReconnect(t *testing.T) {
	var connectCount atomic.Int32
	var serverConns sync.WaitGroup

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		count := connectCount.Add(1)

		if count == 1 {
			// First connection: send welcome then close.
			sendEnvelope(t, conn, makeWelcome("sess-1"))
			time.Sleep(50 * time.Millisecond)
			conn.Close()
			return
		}

		// Second connection: send welcome and keep open.
		sendEnvelope(t, conn, makeWelcome("sess-2"))
		serverConns.Add(1)
		defer serverConns.Done()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client := NewClient(wsURL, "tok")

	var onConnectCalls atomic.Int32
	done := make(chan struct{})

	go func() {
		client.connectWithReconnect(func() {
			n := onConnectCalls.Add(1)
			if n >= 2 {
				// After 2 successful connections, stop by closing.
				client.Close()
			}
		}, func(d time.Duration) {
			// Use a tiny sleep in tests.
			time.Sleep(10 * time.Millisecond)
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for reconnect loop")
	}

	if got := onConnectCalls.Load(); got < 2 {
		t.Errorf("onConnect called %d times, want >= 2", got)
	}

	if got := connectCount.Load(); got < 2 {
		t.Errorf("server saw %d connections, want >= 2", got)
	}
}

func TestClose_BeforeConnect(t *testing.T) {
	client := NewClient("ws://localhost:0", "tok")
	if err := client.Close(); err != nil {
		t.Errorf("Close() on unconnected client: %v", err)
	}
}

func TestSend_NotConnected(t *testing.T) {
	client := NewClient("ws://localhost:0", "tok")
	env := protocol.NewEnvelope(protocol.TypeDeviceHeartbeat, "dev-1")
	err := client.Send(env)
	if err == nil {
		t.Fatal("Send() on unconnected client should return error")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("Send() error = %q, want 'not connected'", err.Error())
	}
}
