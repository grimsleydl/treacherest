package treacherest

import (
	"embed"
	_ "embed"
)

// Embed the treachery cards JSON file
//
//go:embed static/treachery-cards.json
var TreacheryCardsJSON []byte

// Embed all card images
//
//go:embed static/images/cards/*.jpg
var CardImagesFS embed.FS
