package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RoomQRCode serves a normal PNG image for the room join QR code.
func (h *Handler) RoomQRCode(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "code")
	if _, err := h.store.GetRoom(roomCode); err != nil {
		http.NotFound(w, r)
		return
	}

	qrURL := fmt.Sprintf("%s/room/%s", getBaseURL(r), roomCode)
	encodedPNG, err := generateQRCode(qrURL)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	png, err := base64.StdEncoding.DecodeString(encodedPNG)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}
