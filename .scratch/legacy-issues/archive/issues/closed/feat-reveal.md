## Feature: per‑player “Reveal” + always‑clickable Leader row

> **Goal**
>
> * Once the game is *playing*, each player can reveal *their own* role inline by expanding their own row.
> * The **Leader** row is always expandable for everyone (mini role card + oracle text).
> * Keep it **private by default** (your reveal doesn’t expose your role to others) unless you later decide to wire it into the existing `RoleRevealed` state.

### Why this approach fits your codebase

* The game page already shows (1) your full role card at the top and (2) a “Players” list below. We’ll enhance the Players list with a lightweight **client‑side collapsible** per row that can show a **mini role card**. This avoids new SSE signals and avoids modifying the whitelist in `ValidateSSERequest`. The StreamGame loop sometimes sends full re-renders (e.g., on `game_playing`), so the collapsible state is ephemeral, which is acceptable for a “peek” UI.
* The Leader is already globally revealed after the countdown (`LeaderRevealed = true`), but the current UI only shows a banner + name. We’ll add a compact mini card in the Leader’s row so everyone can expand it anytime.

---

## Implementation plan (drop‑in changes)

### 1) Add a **MiniRoleCard** partial (Templ)

Put this *in the same package* as your other page helpers so we can reuse `FormatCardText`. For example, append to `nix/app/internal/views/pages/game.templ` (below existing helpers like `FormatCardText`), or add a new `pages/_partials.templ` in the `pages` package.

```templ
// nix/app/internal/views/pages/game.templ  (append near helpers)

templ MiniRoleCard(card *game.Card) {
	if card == nil {
		<div class="text-sm opacity-60">No role assigned.</div>
		return
	}
	<div class="border rounded-lg p-3 bg-base-100 space-y-2">
		<div class="flex items-center gap-2">
			<strong class="truncate">{ card.Name }</strong>
			<span class="badge badge-outline">{ string(card.GetRoleType()) }</span>
			if card.URI != "" {
				<a href={ templ.SafeURL(card.URI) } target="_blank" rel="noopener noreferrer"
				   class="link link-primary text-xs" title="View on Oracle">
					Oracle
				</a>
			}
		</div>
		<div class="text-sm leading-relaxed">
			@FormatCardText(card.Text)
		</div>
	</div>
}
```

This uses your existing `Card` fields and helpers (`GetRoleType`, `FormatCardText`).

---

### 2) Make player rows expandable and show mini cards

Edit the **Players** card in `GameContent` (`nix/app/internal/views/pages/game.templ`). Replace the simple list with DaisyUI collapsibles:

```templ
// Inside GameContent → the "Players" card section
<div class="card bg-base-200 shadow-lg p-6 max-w-md w-full">
	<h2 class="text-xl font-bold mb-4 text-primary">Players</h2>
	<div class="join join-vertical w-full">

		// Only allow reveal once the game is actually playing
		for _, p := range room.GetActivePlayers() {
			// Unique id so label/button can toggle checkbox accessibly
			id := fmt.Sprintf("reveal-%s", p.ID)

			<div class="collapse collapse-arrow join-item border border-base-300">
				<input
					type="checkbox"
					id={ id }
					disabled?={ room.State != game.StatePlaying }   // reveal only after start
				/>
				<div class="collapse-title flex items-center justify-between gap-2">
					<div class="flex items-center gap-2">
						<span class={ templ.KV("font-bold", p.ID == currentPlayer.ID) }>{ p.Name }</span>
						if p.ID == currentPlayer.ID {
							<span class="badge badge-primary badge-xs">You</span>
						}
						if p.Role != nil && p.Role.GetRoleType() == game.RoleLeader {
							<span class="badge badge-warning badge-xs">Leader</span>
						}
						if p.RoleRevealed {
							<span class="badge badge-ghost badge-xs">{ p.Role.Name }</span>
						}
					</div>
					// Button improves a11y discoverability vs. clicking title area
					<label for={ id } class="btn btn-ghost btn-xs"
					       aria-controls={ id + "-panel" }>
						Reveal
					</label>
				</div>

				<div id={ id + "-panel" } class="collapse-content">
					// Who can see a mini card here?
					// - Always show Leader's mini card to everyone
					// - Show your own mini card to you
					// - If RoleRevealed is true (global), show to everyone
					if (p.Role != nil && p.Role.GetRoleType() == game.RoleLeader) ||
					   (p.ID == currentPlayer.ID && p.Role != nil) ||
					   (p.RoleRevealed && p.Role != nil) {
						@MiniRoleCard(p.Role)
					} else {
						<div class="text-sm opacity-60">Unknown role.</div>
					}
				</div>
			</div>
		}
	</div>
</div>
```

* We keep **Leader** always expandable for every player (whether or not the current player is the leader). This aligns with your `LeaderRevealed` semantics at countdown completion, but provides an actual card preview instead of only a banner.
* We **don’t** mutate `Player.RoleRevealed` on click here—so this is a *private peek* UI. The global `RoleRevealed` badge remains useful for future mechanics (e.g., official reveals on death); if it’s already true, everyone will see the mini card on expand. 

> **Note:** StreamGame sometimes re-renders the game container (e.g., on `game_playing`), which will collapse open rows; that’s OK for a glance UI. If you ever want to persist open/closed across patches, you’d need either client persistence (localStorage keyed by player id) or to add a whitelisted Datastar signal—today your validator only expects known signals (e.g., `countdown`). 

---

### 3) (Optional) Add a server endpoint to *globally* reveal a role

If you *do* want a button that flips the public `RoleRevealed` bit (versus the private peek above), wire an endpoint that only the **owner of that player row** (or a host/leader if you want) can call:

```go
// handlers/actions_reveal.go (new)
func (h *Handler) ToggleReveal(w http.ResponseWriter, r *http.Request) {
    roomCode := chi.URLParam(r, "code")
    playerID := chi.URLParam(r, "playerID")

    room, err := h.store.GetRoom(roomCode)
    if err != nil { http.Error(w, "Room not found", http.StatusNotFound); return }

    // auth: only the same player (cookie) OR host/leader can toggle
    cookie, err := r.Cookie("player_" + roomCode)
    if err != nil { http.Error(w, "Not in room", http.StatusUnauthorized); return }
    me := room.GetPlayer(cookie.Value)
    if me == nil { http.Error(w, "Player not found", http.StatusUnauthorized); return }

    target := room.GetPlayer(playerID)
    if target == nil { http.Error(w, "Target not found", http.StatusBadRequest); return }

    // simple rule: player can reveal themselves; expand for host/leader later if desired
    canToggle := me.ID == target.ID
    if !canToggle {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    target.RoleRevealed = !target.RoleRevealed
    h.store.UpdateRoom(room)

    // Push a game update so everyone gets the updated badge/mini card logic
    h.eventBus.Publish(Event{ Type: "player_updated", RoomCode: room.Code, Data: room })
}
```

* Route: `r.Post("/room/{code}/reveal/{playerID}", h.ToggleReveal)` (where you register routes). Then add a tiny button in the row to call `@post('/room/{code}/reveal/{playerID}')`. Your event bus + StreamGame already re-renders the page for unknown event types in the default branch, so the badge/UI will refresh.

---

## Small repo improvements (while you’re here)

1. **Stable player order.** `GetActivePlayers()` returns a map copy without sorting, so the UI order may jump between SSE patches. Consider sorting by `JoinedAt`. 

   ```go
   func (r *Room) GetActivePlayers() []*Player {
       r.mu.RLock(); defer r.mu.RUnlock()
       players := make([]*Player, 0, len(r.Players))
       for _, p := range r.Players { if !p.IsHost { players = append(players, p) } }
       sort.Slice(players, func(i, j int) bool { return players[i].JoinedAt.Before(players[j].JoinedAt) })
       return players
   }
   ```

2. **Leader exposure is already global; surface the card.** After countdown you set `LeaderRevealed = true` and you show a banner; now we’ll also show the mini card under the Leader’s row for a better UX. (No backend change required.)

3. **A11y polish.** In the collapsible list, connect the “Reveal” `<label>` to its `<input>` and add `aria-controls` on the panel (shown above). Also ensure the “disabled before playing” state is announced with `aria-disabled` if you keep that rule. The page is already good about live regions (“Revealing roles in…”, etc.). 

4. **Keep “private reveal” client‑only.** Because StreamGame often updates `#game-container` on major events, avoid storing per‑row open state in Datastar signals unless you expand your SSE validator whitelist (you currently gate which signals can be patched). The current approach avoids that entirely. 

---

## Minimal code diff (reference)

**A) Add MiniRoleCard** (see snippet in Step 1) in `pages` package. Uses your existing helpers. 

**B) Replace Players list** in `GameContent` with the collapsible version (Step 2). This augments your current “leader banner + players” block. 

*(Optional C) Public reveal endpoint + button* if you want “reveal to everyone” semantics later (Step 3). Uses your existing event bus + StreamGame’s default “full re-render” behavior. 

---

## Test ideas

* Extend `tests/playwright/sse-countdown-sync.spec.ts`: after countdown completes and roles appear (you already assert the main role card is visible), programmatically click the “Reveal” control for the current player and assert the mini card shows under their row; then click the Leader row and assert the mini card shows there for a non‑leader browser too. You already have a pattern for multi‑browser flows and assertions.

---

If you want, I can turn the above into concrete PR‑ready patches (templ + small Go route) and wire the optional public reveal; but with the snippets here you can drop in the “private peek + always‑expandable leader” solution immediately without touching the backend.

