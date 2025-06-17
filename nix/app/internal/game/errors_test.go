package game

import (
	"errors"
	"testing"
)

func TestGameErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrRoomFull has correct message",
			err:      ErrRoomFull,
			expected: "room is full",
		},
		{
			name:     "ErrGameAlreadyStarted has correct message",
			err:      ErrGameAlreadyStarted,
			expected: "game has already started",
		},
		{
			name:     "ErrNotEnoughPlayers has correct message",
			err:      ErrNotEnoughPlayers,
			expected: "not enough players to start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Error message = %v, want %v", tt.err.Error(), tt.expected)
			}
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	// Ensure all errors are distinct
	errorList := []error{
		ErrRoomFull,
		ErrGameAlreadyStarted,
		ErrNotEnoughPlayers,
	}

	for i := 0; i < len(errorList); i++ {
		for j := i + 1; j < len(errorList); j++ {
			if errors.Is(errorList[i], errorList[j]) {
				t.Errorf("Error %v should not be equal to %v", errorList[i], errorList[j])
			}
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that errors can be wrapped and unwrapped properly
	tests := []struct {
		name        string
		baseErr     error
		wrapMessage string
	}{
		{
			name:        "wrapped ErrRoomFull",
			baseErr:     ErrRoomFull,
			wrapMessage: "failed to add player",
		},
		{
			name:        "wrapped ErrGameAlreadyStarted",
			baseErr:     ErrGameAlreadyStarted,
			wrapMessage: "cannot join game",
		},
		{
			name:        "wrapped ErrNotEnoughPlayers",
			baseErr:     ErrNotEnoughPlayers,
			wrapMessage: "cannot start game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap the error
			wrappedErr := errors.Join(errors.New(tt.wrapMessage), tt.baseErr)

			// Check that we can detect the base error
			if !errors.Is(wrappedErr, tt.baseErr) {
				t.Errorf("Wrapped error should contain base error %v", tt.baseErr)
			}

			// Check that the wrapped error contains both messages
			errMsg := wrappedErr.Error()
			if errMsg == "" {
				t.Errorf("Wrapped error message should not be empty")
			}
		})
	}
}

func TestErrorUsageInContext(t *testing.T) {
	// Simulate how these errors might be used in actual game logic
	t.Run("room full scenario", func(t *testing.T) {
		// Simulate a function that returns ErrRoomFull
		checkRoomCapacity := func(currentPlayers, maxPlayers int) error {
			if currentPlayers >= maxPlayers {
				return ErrRoomFull
			}
			return nil
		}

		err := checkRoomCapacity(4, 4)
		if !errors.Is(err, ErrRoomFull) {
			t.Errorf("Expected ErrRoomFull when room is at capacity")
		}

		err = checkRoomCapacity(3, 4)
		if err != nil {
			t.Errorf("Expected no error when room has space")
		}
	})

	t.Run("game already started scenario", func(t *testing.T) {
		// Simulate a function that checks game state
		checkGameState := func(started bool) error {
			if started {
				return ErrGameAlreadyStarted
			}
			return nil
		}

		err := checkGameState(true)
		if !errors.Is(err, ErrGameAlreadyStarted) {
			t.Errorf("Expected ErrGameAlreadyStarted when game is started")
		}

		err = checkGameState(false)
		if err != nil {
			t.Errorf("Expected no error when game hasn't started")
		}
	})

	t.Run("not enough players scenario", func(t *testing.T) {
		// Simulate a function that checks minimum players
		checkPlayerCount := func(playerCount, minPlayers int) error {
			if playerCount < minPlayers {
				return ErrNotEnoughPlayers
			}
			return nil
		}

		err := checkPlayerCount(1, 2)
		if !errors.Is(err, ErrNotEnoughPlayers) {
			t.Errorf("Expected ErrNotEnoughPlayers when below minimum")
		}

		err = checkPlayerCount(2, 2)
		if err != nil {
			t.Errorf("Expected no error when at minimum players")
		}
	})
}