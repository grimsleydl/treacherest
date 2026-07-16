package components

import (
	"treacherest/internal/game"

	"github.com/a-h/templ"
)

// UndercoverUnveilModal builds the owner-only Undercover unveil flow. The
// advisory is omitted once public state shows another identity was revealed.
func UndercoverUnveilModal(room *game.Room, player *game.Player) templ.Component {
	var advisory templ.Component = templ.NopComponent
	if !game.HasOtherRevealedIdentity(room, player) {
		advisory = undercoverUnveilAdvisory()
	}
	return undercoverUnveilModal(room, player, advisory)
}
