package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"treacherest/internal/game"
	"treacherest/internal/store"
	
	"github.com/go-chi/chi/v5"
)

// BenchmarkRoomCreation benchmarks the time to create a new room
func BenchmarkRoomCreation(b *testing.B) {
	s := store.NewMemoryStore()
	h := New(s)

	// Create a form request
	form := url.Values{}
	form.Add("playerName", "TestPlayer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateRoom(w, req)

		if w.Code != http.StatusSeeOther {
			b.Fatalf("expected status %d, got %d", http.StatusSeeOther, w.Code)
		}
	}

	b.ReportMetric(float64(b.Elapsed().Milliseconds())/float64(b.N), "ms/op")
}

// BenchmarkJoinRoom benchmarks the time to join an existing room
func BenchmarkJoinRoom(b *testing.B) {
	s := store.NewMemoryStore()
	h := New(s)

	// Create a room first
	room, _ := s.CreateRoom()
	roomCode := room.Code

	// Set up chi router
	router := chi.NewRouter()
	router.Get("/room/{code}", h.JoinRoom)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a unique player name for each iteration
		playerName := fmt.Sprintf("Player%d", i)
		
		req := httptest.NewRequest("GET", "/room/"+roomCode+"?name="+playerName, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Clean up the player for next iteration
		// Find player by name
		for _, p := range room.GetPlayers() {
			if p.Name == playerName {
				room.RemovePlayer(p.ID)
				break
			}
		}
	}

	b.ReportMetric(float64(b.Elapsed().Milliseconds())/float64(b.N), "ms/op")
}

// BenchmarkSSEBroadcast benchmarks the time to broadcast to N clients
func BenchmarkSSEBroadcast(b *testing.B) {
	testCases := []int{1, 5, 10, 20, 50}

	for _, numClients := range testCases {
		b.Run(fmt.Sprintf("%d_clients", numClients), func(b *testing.B) {
			s := store.NewMemoryStore()
			h := New(s)

			// Create a room with players
			room, _ := s.CreateRoom()
			for i := 0; i < numClients; i++ {
				player := game.NewPlayer(fmt.Sprintf("player%d", i), fmt.Sprintf("Player %d", i), fmt.Sprintf("session%d", i))
				room.AddPlayer(player)
			}
			s.UpdateRoom(room)

			// Subscribe clients
			subscribers := make([]chan Event, numClients)
			for i := 0; i < numClients; i++ {
				subscribers[i] = h.eventBus.Subscribe(room.Code)
			}

			// Measure broadcast time
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				start := time.Now()
				h.eventBus.Publish(Event{
					Type:     "player_joined",
					RoomCode: room.Code,
					Data:     room,
				})
				
				// Wait for all subscribers to receive
				for j := 0; j < numClients; j++ {
					select {
					case <-subscribers[j]:
						// Received
					case <-time.After(100 * time.Millisecond):
						b.Fatalf("subscriber %d timeout", j)
					}
				}
				
				elapsed := time.Since(start)
				b.ReportMetric(float64(elapsed.Microseconds()), "μs/broadcast")
			}

			// Cleanup
			for _, ch := range subscribers {
				h.eventBus.Unsubscribe(room.Code, ch)
			}
		})
	}
}

// BenchmarkConcurrentSSEClients benchmarks handling many concurrent SSE connections
func BenchmarkConcurrentSSEClients(b *testing.B) {
	testCases := []int{10, 50, 100, 200}

	for _, numClients := range testCases {
		b.Run(fmt.Sprintf("%d_concurrent", numClients), func(b *testing.B) {
			s := store.NewMemoryStore()
			h := New(s)

			// Create multiple rooms
			numRooms := numClients / 10 // Average 10 clients per room
			if numRooms < 1 {
				numRooms = 1
			}

			rooms := make([]*game.Room, numRooms)
			for i := 0; i < numRooms; i++ {
				room, _ := s.CreateRoom()
				// Add a player to each room
				player := game.NewPlayer("host"+room.Code, "Host", "session"+room.Code)
				room.AddPlayer(player)
				s.UpdateRoom(room)
				rooms[i] = room
			}

			b.ResetTimer()
			
			// Run concurrent SSE connections
			var wg sync.WaitGroup
			connectionsPerIteration := numClients

			for i := 0; i < b.N; i++ {
				// Create goroutines for concurrent connections
				for j := 0; j < connectionsPerIteration; j++ {
					wg.Add(1)
					go func(clientID int) {
						defer wg.Done()
						
						// Select a room
						room := rooms[clientID%numRooms]
						
						// Subscribe
						ch := h.eventBus.Subscribe(room.Code)
						defer h.eventBus.Unsubscribe(room.Code, ch)
						
						// Simulate SSE connection lifetime
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
						defer cancel()
						
						select {
						case <-ctx.Done():
							// Connection closed
						case <-ch:
							// Received event
						}
					}(j)
				}
				
				wg.Wait()
			}

			b.ReportMetric(float64(numClients), "clients/iteration")
		})
	}
}

// BenchmarkMemoryPerRoom benchmarks memory usage with N players
func BenchmarkMemoryPerRoom(b *testing.B) {
	testCases := []int{2, 4, 8, 16}

	for _, numPlayers := range testCases {
		b.Run(fmt.Sprintf("%d_players", numPlayers), func(b *testing.B) {
			// Force GC before measurement
			runtime.GC()
			runtime.GC()
			
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			s := store.NewMemoryStore()
			h := New(s)

			rooms := make([]*game.Room, b.N)
			
			b.ResetTimer()
			
			// Create b.N rooms with numPlayers each
			for i := 0; i < b.N; i++ {
				room, _ := s.CreateRoom()
				
				// Add players
				for j := 0; j < numPlayers; j++ {
					player := game.NewPlayer(
						fmt.Sprintf("player_%d_%d", i, j),
						fmt.Sprintf("Player %d", j),
						fmt.Sprintf("session_%d_%d", i, j),
					)
					room.AddPlayer(player)
					
					// Subscribe to SSE (simulating real connections)
					ch := h.eventBus.Subscribe(room.Code)
					// Keep channel alive but drain events
					go func(ch chan Event) {
						for range ch {
							// Drain events
						}
					}(ch)
				}
				
				s.UpdateRoom(room)
				rooms[i] = room
			}

			// Force GC and read memory stats
			runtime.GC()
			runtime.GC()
			
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			// Calculate memory per room
			totalMemory := m2.Alloc - m1.Alloc
			memoryPerRoom := float64(totalMemory) / float64(b.N)
			
			b.ReportMetric(memoryPerRoom/1024, "KB/room")
			b.ReportMetric(memoryPerRoom/1024/float64(numPlayers), "KB/player")
		})
	}
}


// Additional benchmarks for specific operations

// BenchmarkEventBusPublish measures raw event bus performance
func BenchmarkEventBusPublish(b *testing.B) {
	eb := NewEventBus()
	
	// Create subscribers
	numSubscribers := 100
	roomCode := "TEST1"
	
	for i := 0; i < numSubscribers; i++ {
		ch := eb.Subscribe(roomCode)
		// Drain events in background
		go func(ch chan Event) {
			for range ch {
			}
		}(ch)
	}

	event := Event{
		Type:     "test_event",
		RoomCode: roomCode,
		Data:     "test data",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Publish(event)
	}

	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N), "ns/publish")
}

// BenchmarkStoreOperations measures store performance
func BenchmarkStoreOperations(b *testing.B) {
	b.Run("CreateRoom", func(b *testing.B) {
		s := store.NewMemoryStore()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := s.CreateRoom()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GetRoom", func(b *testing.B) {
		s := store.NewMemoryStore()
		room, _ := s.CreateRoom()
		code := room.Code
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := s.GetRoom(code)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("UpdateRoom", func(b *testing.B) {
		s := store.NewMemoryStore()
		room, _ := s.CreateRoom()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := s.UpdateRoom(room)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Helper function to simulate realistic room state
func createRealisticRoom(s *store.MemoryStore, numPlayers int) *game.Room {
	room, _ := s.CreateRoom()
	
	for i := 0; i < numPlayers; i++ {
		player := game.NewPlayer(
			fmt.Sprintf("player%d", i),
			fmt.Sprintf("Player %d", i),
			fmt.Sprintf("session%d", i),
		)
		
		// Assign roles if enough players
		if i == 0 && numPlayers >= 3 {
			player.Role = game.LeaderRole
		} else {
			player.Role = game.GuardianRole
		}
		
		room.AddPlayer(player)
	}
	
	// Set game state
	if numPlayers >= 3 {
		room.State = game.StatePlaying
		room.StartedAt = time.Now()
	}
	
	s.UpdateRoom(room)
	return room
}