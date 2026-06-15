package game

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
)

// CoupRoleImage describes the canonical static image file for a Coup role.
type CoupRoleImage struct {
	Role       RoleType
	ID         int
	Slug       string
	StaticPath string
	PublicPath string
}

var coupRoleImageManifest = []CoupRoleImage{
	{Role: RoleKing, ID: 1001, Slug: "king", StaticPath: "static/images/coup/king.jpg", PublicPath: "/static/images/coup/king.jpg"},
	{Role: RoleBlueKnight, ID: 1002, Slug: "blue-knight", StaticPath: "static/images/coup/blue-knight.jpg", PublicPath: "/static/images/coup/blue-knight.jpg"},
	{Role: RoleBlackKnight, ID: 1003, Slug: "black-knight", StaticPath: "static/images/coup/black-knight.jpg", PublicPath: "/static/images/coup/black-knight.jpg"},
	{Role: RoleRedKnight, ID: 1004, Slug: "red-knight", StaticPath: "static/images/coup/red-knight.jpg", PublicPath: "/static/images/coup/red-knight.jpg"},
	{Role: RoleGreenKnight, ID: 1005, Slug: "green-knight", StaticPath: "static/images/coup/green-knight.jpg", PublicPath: "/static/images/coup/green-knight.jpg"},
	{Role: RoleWasteland, ID: 1006, Slug: "wasteland-knight", StaticPath: "static/images/coup/wasteland-knight.jpg", PublicPath: "/static/images/coup/wasteland-knight.jpg"},
}

var coupRoleImageExtensions = []string{".jpg", ".jpeg", ".png", ".webp"}

// CoupRoleImageManifest returns the canonical role-image targets for Coup.
func CoupRoleImageManifest() []CoupRoleImage {
	manifest := make([]CoupRoleImage, len(coupRoleImageManifest))
	copy(manifest, coupRoleImageManifest)
	return manifest
}

// SupportedCoupRoleImageExtensions returns the extensions accepted by the Coup image importer.
func SupportedCoupRoleImageExtensions() []string {
	extensions := make([]string, len(coupRoleImageExtensions))
	copy(extensions, coupRoleImageExtensions)
	return extensions
}

// LoadCoupRoleImages attaches embedded role images to Coup role card templates.
// Missing images are allowed so core Coup gameplay is not blocked before art exists.
func LoadCoupRoleImages(imagesFS fs.FS) error {
	if imagesFS == nil {
		return nil
	}

	for _, roleImage := range coupRoleImageManifest {
		publicPath, imageData, err := readCoupRoleImage(imagesFS, roleImage)
		if err != nil {
			if isMissingImageError(err) {
				continue
			}
			return fmt.Errorf("failed to read Coup role image for %s: %w", roleImage.Role, err)
		}

		card := coupRoleCards[roleImage.Role]
		if card == nil {
			continue
		}
		mimeType := http.DetectContentType(imageData)
		card.Base64Image = fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))
		if publicPath == "" {
			publicPath = roleImage.PublicPath
		}
		card.ImagePath = publicPath
	}

	return nil
}

func readCoupRoleImage(imagesFS fs.FS, roleImage CoupRoleImage) (string, []byte, error) {
	var lastMissing error
	for _, extension := range coupRoleImageExtensions {
		staticPath := fmt.Sprintf("static/images/coup/%s%s", roleImage.Slug, extension)
		imageData, err := fs.ReadFile(imagesFS, staticPath)
		if err == nil {
			return fmt.Sprintf("/static/images/coup/%s%s", roleImage.Slug, extension), imageData, nil
		}
		if !isMissingImageError(err) {
			return "", nil, err
		}
		lastMissing = err
	}
	if lastMissing != nil {
		return "", nil, lastMissing
	}
	return "", nil, fs.ErrNotExist
}

func isMissingImageError(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
