package game

import "errors"

var (
	ErrRoomFull           = errors.New("room is full")
	ErrGameAlreadyStarted = errors.New("game has already started")
	ErrNotEnoughPlayers   = errors.New("not enough players to start")
	ErrDuplicateName      = errors.New("a player with that name already exists in the room")
)
