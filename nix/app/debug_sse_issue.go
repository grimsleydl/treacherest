package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	// Use system Chrome if available
	path, exists := launcher.LookPath()
	if !exists {
		log.Fatal("Chrome not found in system")
	}

	// Launch browser with devtools
	url := launcher.New().Bin(path).Headless(false).Devtools(true).MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	baseURL := "http://localhost:7331"

	// Monitor console for errors
	setupMonitor := func(page *rod.Page, name string) {
		page.MustEval(`(name) => {
			window.playerName = name;
			window.errorLog = [];
			
			// Capture console errors
			window.addEventListener('error', (e) => {
				window.errorLog.push({
					time: new Date().toISOString(),
					player: name,
					message: e.message,
					stack: e.error?.stack
				});
				console.error(name + ' ERROR:', e.message);
			});
			
			// Monitor DOM changes
			const observer = new MutationObserver((mutations) => {
				mutations.forEach((mutation) => {
					if (mutation.type === 'childList' && mutation.removedNodes.length > 0) {
						mutation.removedNodes.forEach(node => {
							if (node.id === 'lobby-container' || node.id === 'lobby-content') {
								console.error(name + ' REMOVED:', node.id);
							}
						});
					}
				});
			});
			
			setTimeout(() => {
				const target = document.getElementById('lobby-container');
				if (target) {
					observer.observe(target.parentElement, { childList: true, subtree: true });
				}
			}, 500);
		}`, name)
	}

	// Player 1 creates room
	fmt.Println("=== Player 1 creating room ===")
	page1 := browser.MustPage()
	
	page1.MustNavigate(baseURL)
	page1.MustElement("input[name='name']").MustInput("Player1")
	page1.MustElement("button[type='submit']").MustClick()
	page1.MustWaitLoad()
	
	roomCode := page1.MustInfo().URL[len(baseURL+"/room/"):]
	setupMonitor(page1, "Player1")
	
	fmt.Printf("Room created: %s\n", roomCode)
	time.Sleep(1 * time.Second)
	
	// Check DOM structure
	printDOM(page1, "Player1 initial")
	
	// Player 2 joins
	fmt.Println("\n=== Player 2 joining ===")
	page2 := browser.MustPage()
	
	page2.MustNavigate(baseURL + "/room/" + roomCode)
	page2.MustElement("input[name='name']").MustInput("Player2")
	page2.MustElement("button[type='submit']").MustClick()
	
	setupMonitor(page2, "Player2")
	time.Sleep(2 * time.Second)
	
	printDOM(page1, "Player1 after P2 joins")
	printDOM(page2, "Player2 after joining")
	
	// Player 3 joins
	fmt.Println("\n=== Player 3 joining ===")
	page3 := browser.MustPage()
	
	page3.MustNavigate(baseURL + "/room/" + roomCode)
	page3.MustElement("input[name='name']").MustInput("Player3")
	page3.MustElement("button[type='submit']").MustClick()
	
	setupMonitor(page3, "Player3")
	time.Sleep(3 * time.Second)
	
	// Check all pages
	for i, page := range []*rod.Page{page1, page2, page3} {
		name := fmt.Sprintf("Player%d", i+1)
		printDOM(page, name+" after P3 joins")
		
		// Check errors
		errors := page.MustEval(`() => window.errorLog`).Arr()
		if len(errors) > 0 {
			fmt.Printf("\n!!! %s ERRORS !!!\n", name)
			for _, err := range errors {
				fmt.Printf("%v\n", err)
			}
		}
	}
	
	fmt.Println("\nPress Enter to close browsers...")
	fmt.Scanln()
}

func printDOM(page *rod.Page, label string) {
	fmt.Printf("\n--- %s DOM ---\n", label)
	
	// Check wrapper div
	hasWrapper := page.MustEval(`() => {
		const wrapper = document.querySelector('[data-on-load*="sse/lobby"]');
		return !!wrapper;
	}`).Bool()
	
	// Check container
	hasContainer := page.MustEval(`() => !!document.getElementById('lobby-container')`).Bool()
	
	// Check content
	hasContent := page.MustEval(`() => !!document.getElementById('lobby-content')`).Bool()
	
	fmt.Printf("Wrapper with data-on-load: %v\n", hasWrapper)
	fmt.Printf("Has #lobby-container: %v\n", hasContainer)
	fmt.Printf("Has #lobby-content: %v\n", hasContent)
	
	// Get HTML structure
	if hasWrapper {
		html := page.MustEval(`() => {
			const wrapper = document.querySelector('[data-on-load*="sse/lobby"]');
			if (!wrapper) return "No wrapper found";
			
			// Get simplified structure
			const getStructure = (el, depth = 0) => {
				const indent = '  '.repeat(depth);
				let str = indent + '<' + el.tagName.toLowerCase();
				if (el.id) str += ' id="' + el.id + '"';
				if (el.hasAttribute('data-on-load')) str += ' data-on-load="..."';
				str += '>';
				
				if (el.children.length > 0 && depth < 3) {
					str += '\n';
					for (const child of el.children) {
						if (child.tagName.toLowerCase() !== 'script') {
							str += getStructure(child, depth + 1) + '\n';
						}
					}
					str += indent;
				}
				str += '</' + el.tagName.toLowerCase() + '>';
				return str;
			};
			
			return getStructure(wrapper);
		}`).Str()
		fmt.Printf("Structure:\n%s\n", html)
	}
}