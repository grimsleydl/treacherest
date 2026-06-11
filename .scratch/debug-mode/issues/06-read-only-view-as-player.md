# Read-Only View As Player

Status: ready-for-agent

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Add a host-only View As Player selector to the Debug Control Surface. The host should be able to select a real player or Debug Player and inspect that player's rendered perspective from setup/lobby contexts and during play.

View As Player is read-only for MVP. It should render what the selected player would see, including private role and information visibility, but must not expose or enable player actions through the host impersonation surface.

## Acceptance criteria

- [ ] Host can select a real player to View As Player in Debug Mode.
- [ ] Host can select a Debug Player to View As Player after Debug Players exist.
- [ ] View As Player is unavailable when Debug Mode is disabled.
- [ ] View As Player is rejected or absent for non-host clients.
- [ ] The rendered perspective includes the selected player's private role and allowed private information.
- [ ] The rendered perspective does not expose other players' hidden information beyond what the selected player should see.
- [ ] The View As Player surface is read-only and does not expose player action controls such as Inquisition or Royal Guard.
- [ ] The selector works in setup/lobby contexts and during play.
- [ ] Tests cover host selection, non-host rejection, selected-player privacy, Debug Player selection, and read-only behavior.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`
- `.scratch/debug-mode/issues/03-start-override-with-debug-players.md`
