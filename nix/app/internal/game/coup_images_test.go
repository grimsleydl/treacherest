package game

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadCoupRoleImages_AttachesEmbeddedImagesToRoleCards(t *testing.T) {
	imageData := []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xff, 0xd9,
	}
	images := fstest.MapFS{}
	for _, role := range CoupRoleImageManifest() {
		images[role.StaticPath] = &fstest.MapFile{Data: imageData}
	}

	if err := LoadCoupRoleImages(images); err != nil {
		t.Fatalf("load Coup role images: %v", err)
	}
	t.Cleanup(clearCoupRoleImagesForTest)

	for _, role := range CoupRoleImageManifest() {
		card := coupRoleCards[role.Role]
		if card.Base64Image == "" {
			t.Fatalf("expected %s to have base64 image", role.Role)
		}
		if !strings.HasPrefix(card.Base64Image, "data:image/jpeg;base64,") {
			t.Fatalf("expected %s jpeg data URI, got %q", role.Role, card.Base64Image)
		}
		if card.ImagePath != role.PublicPath {
			t.Fatalf("expected public path %q for %s, got %q", role.PublicPath, role.Role, card.ImagePath)
		}
	}
}

func TestLoadCoupRoleImages_UsesDetectedImportedExtension(t *testing.T) {
	images := fstest.MapFS{
		"static/images/coup/1001.png": &fstest.MapFile{Data: []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}},
	}

	if err := LoadCoupRoleImages(images); err != nil {
		t.Fatalf("load Coup role images: %v", err)
	}
	t.Cleanup(clearCoupRoleImagesForTest)

	card := coupRoleCards[RoleKing]
	if !strings.HasPrefix(card.Base64Image, "data:image/png;base64,") {
		t.Fatalf("expected png data URI, got %q", card.Base64Image)
	}
	if card.ImagePath != "/static/images/coup/1001.png" {
		t.Fatalf("expected imported png path, got %q", card.ImagePath)
	}
}

func TestLoadCoupRoleImages_AllowsMissingImages(t *testing.T) {
	images := fstest.MapFS{}

	if err := LoadCoupRoleImages(images); err != nil {
		t.Fatalf("expected missing Coup role images to be allowed, got %v", err)
	}
	t.Cleanup(clearCoupRoleImagesForTest)

	players := []*Player{
		NewPlayer("p1", "Player 1", "session-1"),
		NewPlayer("p2", "Player 2", "session-2"),
		NewPlayer("p3", "Player 3", "session-3"),
		NewPlayer("p4", "Player 4", "session-4"),
		NewPlayer("p5", "Player 5", "session-5"),
	}
	if err := AssignCoupRoles(players, CoupPresetFive); err != nil {
		t.Fatalf("assign Coup roles without images: %v", err)
	}
}

func clearCoupRoleImagesForTest() {
	for _, role := range CoupRoleImageManifest() {
		card := coupRoleCards[role.Role]
		card.Base64Image = ""
		card.ImagePath = ""
	}
}
