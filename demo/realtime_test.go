package demo_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
)

// Opt in only against a disposable stack: this test creates users and chats.
func TestRealtimeLocalStack(t *testing.T) {
	base := strings.TrimRight(os.Getenv("CONNEXT_SMOKE_URL"), "/")
	if base == "" {
		t.Skip("set CONNEXT_SMOKE_URL to a disposable local gateway")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	request := func(method, path, token string, body any) map[string]json.RawMessage {
		t.Helper()
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(method, base+path, bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			t.Fatalf("%s %s: status %d: %s", method, path, res.StatusCode, data)
		}
		var result map[string]json.RawMessage
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatal(err)
		}
		if string(result["success"]) != "true" {
			t.Fatalf("%s %s did not succeed: %s", method, path, data)
		}
		return result
	}
	decodeString := func(data json.RawMessage) string {
		t.Helper()
		var value string
		if err := json.Unmarshal(data, &value); err != nil || value == "" {
			t.Fatalf("expected nonempty string, got %s: %v", data, err)
		}
		return value
	}
	newUser := func(label string) (string, string) {
		t.Helper()
		email := fmt.Sprintf("net1-%s-%d@example.com", label, time.Now().UnixNano())
		request("POST", "/users/register", "", map[string]any{
			"username":  "net1-" + label + fmt.Sprint(time.Now().UnixNano()),
			"password":  "SmokeTestPassword123!",
			"contact":   map[string]string{"email": email},
			"education": map[string]string{"university": "Test", "major": "Test"},
		})
		login := request("POST", "/users/login", "", map[string]string{
			"email": email, "password": "SmokeTestPassword123!",
		})
		token := decodeString(login["jwtToken"])
		var user map[string]json.RawMessage
		if err := json.Unmarshal(request("GET", "/users/me", token, nil)["data"], &user); err != nil {
			t.Fatal(err)
		}
		return token, decodeString(user["user_id"])
	}
	token1, id1 := newUser("alice")
	token2, id2 := newUser("bob")
	connect := func(token string) *websocket.Conn {
		t.Helper()
		wsURL := "ws" + strings.TrimPrefix(base, "http") + "/chats/ws/?token=" + url.QueryEscape(token)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		return conn
	}
	waitEvent := func(conn *websocket.Conn, kind, key, value string) {
		t.Helper()
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		for {
			var event struct {
				Type string            `json:"type"`
				Data map[string]string `json:"data"`
			}
			if err := conn.ReadJSON(&event); err != nil {
				t.Fatalf("waiting for %s (%s=%s): %v", kind, key, value, err)
			}
			if event.Type == kind && event.Data[key] == value {
				return
			}
		}
	}
	ws1 := connect(token1)
	waitEvent(ws1, "user_joined", "user_id", id1)
	ws2 := connect(token2)
	waitEvent(ws1, "user_joined", "user_id", id2)
	waitEvent(ws2, "user_joined", "user_id", id2)
	checkPresence := func(expectSecond bool) {
		t.Helper()
		var data struct {
			Users []struct {
				ID string `json:"user_id"`
			} `json:"online_users"`
		}
		if err := json.Unmarshal(request("GET", "/chats/users", token1, nil)["data"], &data); err != nil {
			t.Fatal(err)
		}
		found := map[string]bool{}
		for _, user := range data.Users {
			found[user.ID] = true
		}
		if !found[id1] || found[id2] != expectSecond {
			t.Fatalf("incorrect online users: %+v", found)
		}
	}
	checkPresence(true)
	t.Log("login, WebSocket join broadcasts, and online-user list passed")
	chatID := func(res map[string]json.RawMessage) string {
		t.Helper()
		var data map[string]json.RawMessage
		if err := json.Unmarshal(res["data"], &data); err != nil {
			t.Fatal(err)
		}
		return decodeString(data["chat_id"])
	}
	direct := chatID(request("POST", "/chats/", token1, map[string]string{"recipient_id": id2}))
	group := chatID(request("POST", "/chats/group", token1, map[string]string{"group_name": "NET-1 smoke"}))
	request("POST", "/chats/"+group+"/join", token2, nil)
	for _, chat := range []string{direct, group} {
		for i, token := range []string{token1, token2} {
			message := fmt.Sprintf("net1-%s-direction-%d", chat, i)
			request("POST", "/chats/send", token, map[string]string{"chat_id": chat, "message": message})
			receiver := ws2
			if i == 1 {
				receiver = ws1
			}
			waitEvent(receiver, "chat.gateway", "message", message)
			var history []struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(request("GET", "/chats/"+chat+"/messages", token1, nil)["data"], &history); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, entry := range history {
				found = found || entry.Message == message
			}
			if !found {
				t.Fatal("delivered message missing from MongoDB-backed history")
			}
		}
	}
	t.Log("direct and group chat delivery in both directions and MongoDB history passed")
	ws2.Close()
	waitEvent(ws1, "user_left", "user_id", id2)
	checkPresence(false)
	ws2 = connect(token2)
	waitEvent(ws1, "user_joined", "user_id", id2)
	checkPresence(true)
	t.Log("disconnect and reconnect presence updates passed")
}
