package game

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Backup-related errors
var (
	ErrBackupExpired      = errors.New("backup is too old")
	ErrBackupTampered     = errors.New("backup failed authentication")
	ErrBackupInvalid      = errors.New("backup is invalid")
	ErrRoomCodeMismatch   = errors.New("backup room code doesn't match")
	ErrBackupKeyInvalid   = errors.New("backup encryption key is invalid")
	ErrBackupKeyTooShort  = errors.New("backup encryption key must be 32 bytes (64 hex chars)")
)

const (
	// BackupMaxAge is the maximum age of a backup before it's rejected
	BackupMaxAge = 1 * time.Hour

	// BackupVersion is the current backup schema version
	BackupVersion = 1
)

// StateBackup represents a serialized game state backup
type StateBackup struct {
	Version   int       `json:"v"`
	Timestamp time.Time `json:"ts"`
	RoomCode  string    `json:"code"`
	Room      *Room     `json:"room"`
}

// BackupService handles creating and restoring encrypted game state backups
type BackupService struct {
	encryptionKey     []byte
	encryptionEnabled bool
}

// NewBackupService creates a new backup service
// key should be a 32-byte key for AES-256, provided as a 64-character hex string
// If key is empty and encryption is enabled, an error is returned
func NewBackupService(hexKey string, enabled bool) (*BackupService, error) {
	var key []byte

	if enabled {
		if hexKey == "" {
			return nil, ErrBackupKeyInvalid
		}

		var err error
		key, err = hex.DecodeString(hexKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrBackupKeyInvalid, err)
		}

		if len(key) != 32 {
			return nil, ErrBackupKeyTooShort
		}
	}

	return &BackupService{
		encryptionKey:     key,
		encryptionEnabled: enabled,
	}, nil
}

// CreateBackup serializes and optionally encrypts room state
func (s *BackupService) CreateBackup(room *Room) (string, error) {
	if room == nil {
		return "", errors.New("room is nil")
	}

	backup := StateBackup{
		Version:   BackupVersion,
		Timestamp: time.Now(),
		RoomCode:  room.Code,
		Room:      room,
	}

	plaintext, err := json.Marshal(backup)
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup: %w", err)
	}

	if !s.encryptionEnabled {
		// Debug mode: return plaintext JSON
		return string(plaintext), nil
	}

	ciphertext, err := s.encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt backup: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RestoreBackup decrypts and validates a backup, returns the reconstructed Room
func (s *BackupService) RestoreBackup(data string, expectedRoomCode string) (*Room, error) {
	if data == "" {
		return nil, ErrBackupInvalid
	}

	var plaintext []byte
	var err error

	if s.encryptionEnabled {
		ciphertext, decodeErr := base64.StdEncoding.DecodeString(data)
		if decodeErr != nil {
			return nil, ErrBackupInvalid
		}

		plaintext, err = s.decrypt(ciphertext)
		if err != nil {
			return nil, ErrBackupTampered
		}
	} else {
		// Debug mode: data is plaintext JSON
		plaintext = []byte(data)
	}

	var backup StateBackup
	if err = json.Unmarshal(plaintext, &backup); err != nil {
		return nil, ErrBackupInvalid
	}

	// Validate backup version
	if backup.Version > BackupVersion {
		return nil, fmt.Errorf("%w: backup version %d is newer than supported %d", ErrBackupInvalid, backup.Version, BackupVersion)
	}

	// Validate timestamp - reject if too old
	if time.Since(backup.Timestamp) > BackupMaxAge {
		return nil, ErrBackupExpired
	}

	// Validate room code matches
	if backup.RoomCode != expectedRoomCode {
		return nil, ErrRoomCodeMismatch
	}

	// Validate room is not nil
	if backup.Room == nil {
		return nil, ErrBackupInvalid
	}

	return backup.Room, nil
}

// IsEnabled returns whether encryption is enabled
func (s *BackupService) IsEnabled() bool {
	return s.encryptionEnabled
}

// encrypt uses AES-256-GCM to encrypt plaintext
// Returns: nonce + ciphertext + auth tag (all concatenated)
func (s *BackupService) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Seal: prepends nonce, appends auth tag
	// Output format: nonce (12 bytes) + ciphertext + auth tag (16 bytes)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt uses AES-256-GCM to decrypt ciphertext
// Input format: nonce (12 bytes) + ciphertext + auth tag (16 bytes)
func (s *BackupService) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrBackupInvalid
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Open: verifies auth tag and decrypts
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// GCM authentication failed - data was tampered
		return nil, ErrBackupTampered
	}

	return plaintext, nil
}
