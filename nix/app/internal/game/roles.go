package game

import (
	"math/rand"
)

// RoleType represents the type of role
type RoleType string

const (
	RoleLeader   RoleType = "Leader"
	RoleGuardian RoleType = "Guardian"
	RoleAssassin RoleType = "Assassin"
	RoleTraitor  RoleType = "Traitor"
)

// Role represents a player's role in the game
type Role struct {
	Type         RoleType
	Name         string
	Description  string
	WinCondition string
}

var (
	// Define all roles
	LeaderRole = &Role{
		Type:         RoleLeader,
		Name:         "Leader",
		Description:  "You are the Leader. Your identity is public.",
		WinCondition: "Survive and be the last player standing",
	}

	GuardianRole = &Role{
		Type:         RoleGuardian,
		Name:         "Guardian",
		Description:  "You are a Guardian. Protect the Leader at all costs.",
		WinCondition: "Win or lose with the Leader",
	}

	AssassinRole = &Role{
		Type:         RoleAssassin,
		Name:         "Assassin",
		Description:  "You are an Assassin. Your goal is to eliminate the Leader.",
		WinCondition: "Win if the Leader is eliminated",
	}

	TraitorRole = &Role{
		Type:         RoleTraitor,
		Name:         "Traitor",
		Description:  "You are the Traitor. Trust no one.",
		WinCondition: "Be the last player standing",
	}
)

// AssignRoles assigns roles to players based on player count
func AssignRoles(players []*Player) {
	count := len(players)

	// Shuffle players first
	shuffled := make([]*Player, count)
	copy(shuffled, players)
	rand.Shuffle(count, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Assign roles based on player count
	roles := getRoleDistribution(count)
	for i, role := range roles {
		shuffled[i].Role = role
		// Leader is always revealed
		if role.Type == RoleLeader {
			shuffled[i].RoleRevealed = true
		}
	}
}

// getRoleDistribution returns the role distribution based on player count
func getRoleDistribution(playerCount int) []*Role {
	switch playerCount {
	case 1:
		return []*Role{LeaderRole}
	case 2:
		return []*Role{LeaderRole, TraitorRole}
	case 3:
		return []*Role{LeaderRole, GuardianRole, TraitorRole}
	case 4:
		return []*Role{LeaderRole, GuardianRole, GuardianRole, TraitorRole}
	case 5:
		return []*Role{LeaderRole, GuardianRole, GuardianRole, AssassinRole, TraitorRole}
	case 6:
		return []*Role{LeaderRole, GuardianRole, GuardianRole, AssassinRole, AssassinRole, TraitorRole}
	case 7:
		return []*Role{LeaderRole, GuardianRole, GuardianRole, GuardianRole, AssassinRole, AssassinRole, TraitorRole}
	case 8:
		return []*Role{LeaderRole, GuardianRole, GuardianRole, GuardianRole, AssassinRole, AssassinRole, TraitorRole, TraitorRole}
	default:
		// Fallback for edge cases - just assign Leader role to everyone
		roles := make([]*Role, playerCount)
		for i := range roles {
			if i == 0 {
				roles[i] = LeaderRole
			} else {
				roles[i] = GuardianRole
			}
		}
		return roles
	}
}
