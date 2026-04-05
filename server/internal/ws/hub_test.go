package ws

import (
	"sync"
	"testing"
	"time"
)

func startTestHub(onExpire func(string)) *Hub {
	if onExpire == nil {
		onExpire = func(string) {}
	}
	hub := NewHub(onExpire)
	go hub.Run()
	return hub
}

func TestCreateAndGetRooms(t *testing.T) {
	hub := startTestHub(nil)
	defer hub.Close()

	hub.CreateRoom(&Room{
		ID:        "room1",
		Name:      "Test",
		Clients:   make(map[string]*Client),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	rooms := hub.GetRooms()
	if len(rooms) != 1 {
		t.Fatalf("got %d rooms, want 1", len(rooms))
	}
	if rooms[0].ID != "room1" {
		t.Errorf("room ID = %q, want %q", rooms[0].ID, "room1")
	}
}

func TestGetRoomInfo(t *testing.T) {
	hub := startTestHub(nil)
	defer hub.Close()

	info := hub.GetRoomInfo("nonexistent")
	if info.Exists {
		t.Error("nonexistent room should not exist")
	}

	hub.CreateRoom(&Room{
		ID:        "room1",
		Name:      "Test",
		Clients:   make(map[string]*Client),
		ExpiresAt: time.Now().Add(time.Hour),
		HasPIN:    true,
	})

	info = hub.GetRoomInfo("room1")
	if !info.Exists {
		t.Error("room should exist")
	}
	if !info.HasPIN {
		t.Error("room should have PIN")
	}
	if info.ClientCount != 0 {
		t.Errorf("ClientCount = %d, want 0", info.ClientCount)
	}
}

func TestRegisterAndUnregister(t *testing.T) {
	hub := startTestHub(nil)
	defer hub.Close()

	hub.CreateRoom(&Room{
		ID:        "room1",
		Name:      "Test",
		Clients:   make(map[string]*Client),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	cl := &Client{
		ConnID:   "conn1",
		ID:       "client1",
		RoomID:   "room1",
		Username: "alice",
		Message:  make(chan *Message, 10),
	}

	hub.Register <- cl
	// drain the presence message
	<-cl.Message

	info := hub.GetRoomInfo("room1")
	if info.ClientCount != 1 {
		t.Errorf("after register: ClientCount = %d, want 1", info.ClientCount)
	}

	hub.Unregister <- cl
	// drain system + presence messages from broadcast
	time.Sleep(50 * time.Millisecond)

	info = hub.GetRoomInfo("room1")
	if info.ClientCount != 0 {
		t.Errorf("after unregister: ClientCount = %d, want 0", info.ClientCount)
	}
}

func TestBroadcast(t *testing.T) {
	hub := startTestHub(nil)
	defer hub.Close()

	hub.CreateRoom(&Room{
		ID:        "room1",
		Name:      "Test",
		Clients:   make(map[string]*Client),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	alice := &Client{ConnID: "conn-a", ID: "1", RoomID: "room1", Username: "alice", Message: make(chan *Message, 10)}
	bob := &Client{ConnID: "conn-b", ID: "2", RoomID: "room1", Username: "bob", Message: make(chan *Message, 10)}

	hub.Register <- alice
	hub.Register <- bob
	// wait for hub to finish post-registration sends before draining
	time.Sleep(50 * time.Millisecond)
	// drain presence + countdown messages sent on join
	drainMessages := func(cl *Client) {
		for {
			select {
			case <-cl.Message:
			default:
				return
			}
		}
	}
	drainMessages(alice)
	drainMessages(bob)

	hub.Broadcast <- &Message{
		Content:  "hello",
		RoomID:   "room1",
		Username: "alice",
		Type:     MessageTypeChat,
		SenderID: "conn-a",
	}

	select {
	case msg := <-bob.Message:
		if msg.Content != "hello" {
			t.Errorf("bob got content %q, want %q", msg.Content, "hello")
		}
	case <-time.After(time.Second):
		t.Fatal("bob did not receive message")
	}

	select {
	case <-alice.Message:
		t.Error("alice should not receive her own message")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRoomExpiry(t *testing.T) {
	var expiredID string
	var mu sync.Mutex

	hub := startTestHub(func(roomID string) {
		mu.Lock()
		expiredID = roomID
		mu.Unlock()
	})
	defer hub.Close()

	hub.CreateRoom(&Room{
		ID:        "expiring",
		Name:      "Expires Soon",
		Clients:   make(map[string]*Client),
		ExpiresAt: time.Now().Add(-time.Second),
	})

	// wait for the ticker to fire (30s default, but room is already expired)
	// The ticker runs every 30s — too slow for tests. Instead we use the
	// expireRoom channel directly.
	hub.expireRoom <- "expiring"
	time.Sleep(100 * time.Millisecond)

	info := hub.GetRoomInfo("expiring")
	if info.Exists {
		t.Error("expired room should be removed")
	}

	mu.Lock()
	if expiredID != "expiring" {
		t.Errorf("onRoomExpire callback got %q, want %q", expiredID, "expiring")
	}
	mu.Unlock()
}

func TestPresenceBroadcast(t *testing.T) {
	hub := startTestHub(nil)
	defer hub.Close()

	hub.CreateRoom(&Room{
		ID:        "room1",
		Name:      "Test",
		Clients:   make(map[string]*Client),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	alice := &Client{ConnID: "conn-1", ID: "1", RoomID: "room1", Username: "alice", Message: make(chan *Message, 10)}
	hub.Register <- alice

	select {
	case msg := <-alice.Message:
		if msg.Type != MessageTypePresence {
			t.Errorf("expected presence message, got %q", msg.Type)
		}
		if len(msg.Clients) != 1 {
			t.Errorf("expected 1 client in presence, got %d", len(msg.Clients))
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive presence message")
	}
}

func TestConcurrentAccess(t *testing.T) {
	hub := startTestHub(nil)
	defer hub.Close()

	hub.CreateRoom(&Room{
		ID:        "room1",
		Name:      "Test",
		Clients:   make(map[string]*Client),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.GetRooms()
			hub.GetRoomInfo("room1")
			hub.GetClients("room1")
		}()
	}

	wg.Wait()
}
