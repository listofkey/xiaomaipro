package test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"server/pkg/ws"

	"github.com/gorilla/websocket"
)

func TestWebsocket(t *testing.T) {
	hub := ws.NewHub()
	t.Cleanup(func() {
		_ = hub.Close()
	})

	mux := http.NewServeMux()
	mux.Handle("/ws/connect", hub)

	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/connect"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket failed: %v", err)
	}
	defer conn.Close()

	waitForWebsocketClient(t, hub, 1)

	type message struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	expected := message{
		Type:    "notification",
		Content: "hello client",
	}

	if err := hub.BroadcastJSON(expected); err != nil {
		t.Fatalf("broadcast websocket message failed: %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message failed: %v", err)
	}

	var actual message
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("unmarshal websocket message failed: %v", err)
	}

	if actual != expected {
		t.Fatalf("unexpected websocket message: got %+v want %+v", actual, expected)
	}
}

func waitForWebsocketClient(t *testing.T, hub *ws.Hub, expected int) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if hub.ClientCount() == expected {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected %d websocket client(s), got %d", expected, hub.ClientCount())
}

func TestWebsocketServer(t *testing.T) {
	// if !shouldRunWebsocketServerTest(t) {
	// 	t.Skip("run `go test ./test -run TestWebsocketServer -v` or set RUN_WEBSOCKET_SERVER_TEST=1")
	// }

	hub := ws.NewHub()
	t.Cleanup(func() {
		_ = hub.Close()
	})

	mux := http.NewServeMux()
	mux.Handle("/ws/connect", hub)

	server := &http.Server{
		Addr:    "127.0.0.1:18089",
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	t.Log("websocket server started")
	t.Log("connect url: ws://127.0.0.1:18089/ws/connect")
	t.Log("server will keep running for 60 seconds")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case err := <-errCh:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				t.Fatalf("websocket server stopped unexpectedly: %v", err)
			}
			return
		case <-ticker.C:
			_ = hub.BroadcastJSON(map[string]any{
				"type":      "heartbeat",
				"content":   "websocket test server is alive",
				"clients":   hub.ClientCount(),
				"timestamp": time.Now().Format(time.RFC3339),
			})
		case <-timeout.C:
			return
		}
	}
}

func shouldRunWebsocketServerTest(t *testing.T) bool {
	t.Helper()

	if os.Getenv("RUN_WEBSOCKET_SERVER_TEST") != "" {
		return true
	}

	runFlag := flag.Lookup("test.run")
	if runFlag == nil {
		return false
	}

	pattern := runFlag.Value.String()
	if pattern == "" {
		return false
	}

	matched, err := regexp.MatchString(pattern, t.Name())
	if err != nil {
		return false
	}

	return matched
}

func TestSSE(t *testing.T) {
	type message struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	expected := message{
		Type:    "notification",
		Content: "hello sse client",
	}

	payload, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("marshal sse payload failed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/sse/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		if _, err := w.Write([]byte("event: notification\n")); err != nil {
			return
		}
		if _, err := w.Write([]byte("data: " + string(payload) + "\n\n")); err != nil {
			return
		}
		flusher.Flush()

		<-r.Context().Done()
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/sse/connect", nil)
	if err != nil {
		t.Fatalf("create sse request failed: %v", err)
	}

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("connect sse failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		t.Fatalf("unexpected content type: %q", contentType)
	}

	eventName, data, err := readSSEEvent(resp.Body)
	if err != nil {
		t.Fatalf("read sse event failed: %v", err)
	}

	if eventName != "notification" {
		t.Fatalf("unexpected event name: got %q want %q", eventName, "notification")
	}

	var actual message
	if err := json.Unmarshal([]byte(data), &actual); err != nil {
		t.Fatalf("unmarshal sse data failed: %v", err)
	}

	if actual != expected {
		t.Fatalf("unexpected sse message: got %+v want %+v", actual, expected)
	}
}

func TestSSEServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sse/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if err := writeSSEEvent(w, "connected", map[string]any{
			"type":      "connected",
			"content":   "sse test server connected",
			"timestamp": time.Now().Format(time.RFC3339),
		}); err != nil {
			return
		}
		flusher.Flush()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case tick := <-ticker.C:
				if err := writeSSEEvent(w, "heartbeat", map[string]any{
					"type":      "heartbeat",
					"content":   "sse test server is alive",
					"timestamp": tick.Format(time.RFC3339),
				}); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	})

	server := &http.Server{
		Addr:    "127.0.0.1:18090",
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	t.Log("sse server started")
	t.Log("connect url: http://127.0.0.1:18090/sse/connect")
	t.Log("EventSource example: new EventSource('http://127.0.0.1:18090/sse/connect')")
	t.Log("server will keep running for 60 seconds")

	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("sse server stopped unexpectedly: %v", err)
		}
	case <-timeout.C:
	}
}

func readSSEEvent(r io.Reader) (string, string, error) {
	scanner := bufio.NewScanner(r)

	var eventName string
	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if eventName != "" || len(dataLines) > 0 {
				return eventName, strings.Join(dataLines, "\n"), nil
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	return "", "", io.EOF
}

func writeSSEEvent(w io.Writer, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := io.WriteString(w, "event: "+event+"\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "data: "+string(data)+"\n\n"); err != nil {
		return err
	}

	return nil
}
