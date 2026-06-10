# Autoscaling and Default Game Size Issues

## Problem
When creating a new game room, the default role configuration was using the maximum player count (20) instead of a reasonable starting size. This resulted in:
- 1 Leader
- 9 Guardians
- 8 Assassins
- 2 Traitors

This was confusing for users who would see these large numbers before any players had even joined.

## Root Cause
In `/workspace/nix/app/internal/store/memory.go`, the room creation logic was using:
```go
roleConfig, _ := roleService.CreateFromPreset("standard", s.config.Server.MaxPlayersPerRoom)
```

This used `MaxPlayersPerRoom` (20) as the player count for the preset, resulting in the 20-player configuration.

## Solution Implemented
Combined two approaches:
1. **Start Small**: Begin with a reasonable default (5 players)
2. **Configurable Default**: Added a new config option for the default game size

### Changes Made:
1. Added `defaultGameSize: 5` to `/workspace/nix/app/config/server.yaml`
2. Updated config struct in `/workspace/nix/app/internal/config/config.go` to include `DefaultGameSize` field with validation
3. Changed room creation to use `defaultGameSize` instead of `maxPlayersPerRoom`
4. Added tests to verify the new behavior

### Result
New rooms now start with a 5-player configuration:
- 1 Leader
- 2 Guardians
- 1 Assassin
- 1 Traitor

The room can still scale up to 20 players maximum, and autoscaling will adjust roles as players join.

## Additional UI Improvements Made
During this conversation, we also:
1. Implemented DaisyUI theme switching with Datastar signals
2. Improved role configuration UI with horizontal button layout and accordion sections
3. Fixed theme persistence across page reloads
4. Styled previously unstyled content on home and lobby pages
5. Made player list card width consistent with role configuration section

## Status
✅ Implemented and tested - rooms now start with sensible defaults while maintaining scalability.