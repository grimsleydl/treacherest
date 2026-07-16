package game

// HasOtherRevealedIdentity reports whether public room state shows that an
// identity other than the owner's has been unveiled.
func HasOtherRevealedIdentity(room *Room, owner *Player) bool {
	if room == nil || owner == nil {
		return false
	}

	for _, player := range room.GetPlayers() {
		if player == nil || player.ID == owner.ID || player.Role == nil {
			continue
		}
		if player.RoleRevealed {
			return true
		}
	}

	return false
}
