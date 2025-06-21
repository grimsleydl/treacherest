package main

import (
	"fmt"
	"log"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"treacherest/internal/handlers"
	"treacherest/internal/store"
)

func main() {
	// Setup server
	store := &handlers.InMemoryStore{
		Rooms: make(map[string]*game.Room),
		Mu:    &sync.RWMutex{},
	}
	eventBus := &handlers.EventBus{
		Subscribers: make(map[string][]chan handlers.Event),
		Mu:          &sync.RWMutex{},
	}
	h := &handlers.Handler{
		Store:    store,
		EventBus: eventBus,
	}
	
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	// Launch browser with devtools
	l := launcher.New().
		Headless(false).
		Devtools(true).
		MustLaunch()
	defer l.Kill()

	browser := rod.New().
		ControlURL(l).
		MustConnect().
		SlowMotion(500 * time.Millisecond)
	defer browser.MustClose()

	log.Println("=== Starting SSE Morph Issue Test ===")
	
	// Player 1 creates room
	page1 := browser.MustPage()
	page1.MustNavigate(server.URL)
	page1.MustWaitStable()
	
	// Monitor console messages
	go page1.EachEvent(func(e *proto.RuntimeConsoleAPICalled) {
		for _, arg := range e.Args {
			log.Printf("[CONSOLE] %s", arg.Value)
		}
	})()
	
	// Setup DOM monitoring  
	page1.MustEval(`() => {
		console.log('Setting up DOM monitoring...');
		
		// Monitor for NoTargetsFound errors
		const originalError = console.error;
		console.error = function(...args) {
			if (args[0] && args[0].toString().includes('NoTargetsFound')) {
				console.log('🔴 NoTargetsFound ERROR DETECTED!', args[0]);
				document.body.style.border = '5px solid red';
			}
			originalError.apply(console, args);
		};
		
		// Monitor DOM mutations
		const observer = new MutationObserver((mutations) => {
			mutations.forEach((mutation) => {
				if (mutation.target.id === 'lobby-container') {
					console.log('🟡 Mutation on lobby-container:', mutation.type);
					if (mutation.removedNodes.length > 0) {
						console.log('🔴 lobby-container had nodes removed!');
					}
				}
				// Check if lobby-container was removed
				if (mutation.removedNodes.length > 0) {
					for (let node of mutation.removedNodes) {
						if (node.id === 'lobby-container') {
							console.log('🔴🔴🔴 CRITICAL: lobby-container WAS REMOVED FROM DOM!');
							document.body.style.backgroundColor = 'red';
						}
					}
				}
			});
		});
		observer.observe(document.body, { childList: true, subtree: true });
		
		console.log('DOM monitoring active');
	}`)
	
	// Create room
	log.Println("Player 1 creating room...")
	page1.MustElement("input[name='playerName']").MustInput("Player 1")
	page1.MustElement("button[type='submit']").MustClick()
	page1.MustWaitStable()
	
	// Get room code
	roomCode := strings.TrimSpace(page1.MustElement(".room-code").MustText())
	log.Printf("Room created: %s", roomCode)
	
	// Verify initial structure
	page1.MustEval(`() => {
		const container = document.querySelector('#lobby-container');
		console.log('Initial check - lobby-container exists:', container !== null);
		if (container) {
			console.log('lobby-container HTML:', container.outerHTML.substring(0, 150) + '...');
		}
	}`)
	
	// Wait a bit to see SSE is connected
	time.Sleep(2 * time.Second)
	
	// Player 2 joins
	log.Println("Player 2 joining...")
	page2 := browser.MustPage()
	page2.MustNavigate(fmt.Sprintf("%s/room/%s?name=Player+2", server.URL, roomCode))
	page2.MustWaitStable()
	
	// Check player 1's DOM after player 2 joins
	time.Sleep(2 * time.Second)
	page1.MustEval(`() => {
		const container = document.querySelector('#lobby-container');
		console.log('After player 2 - lobby-container exists:', container !== null);
		if (!container) {
			console.log('🔴 ERROR: lobby-container missing after player 2 joined!');
			console.log('Current body HTML:', document.body.innerHTML.substring(0, 500));
		}
	}`)
	
	// Player 3 joins - this should trigger the error
	log.Println("Player 3 joining...")
	page3 := browser.MustPage()
	page3.MustNavigate(fmt.Sprintf("%s/room/%s?name=Player+3", server.URL, roomCode))
	page3.MustWaitStable()
	
	// Final check
	time.Sleep(2 * time.Second)
	page1.MustEval(`() => {
		const container = document.querySelector('#lobby-container');
		console.log('After player 3 - lobby-container exists:', container !== null);
		if (!container) {
			console.log('🔴🔴🔴 CONFIRMED: lobby-container is MISSING!');
		}
		
		// Count players visible
		const players = document.querySelectorAll('.player');
		console.log('Number of players visible:', players.length);
	}`)
	
	log.Println("Test complete. Check browser console for details.")
	log.Println("Press Ctrl+C to exit...")
	
	// Keep browser open
	select {}
}