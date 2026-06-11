package game

// CoupInquisitionResultPolicy controls who sees a successful Inquisition result.
type CoupInquisitionResultPolicy string

const (
	CoupInquisitionResultPublic  CoupInquisitionResultPolicy = "public"
	CoupInquisitionResultPrivate CoupInquisitionResultPolicy = "private"
)

// CoupInquisitionState tracks public/default Inquisition flow state.
type CoupInquisitionState struct {
	Attempts  map[string]CoupInquisitionAttempt
	Pending   *CoupInquisitionAttempt
	Last      *CoupInquisitionAttempt
	Succeeded bool
}

// CoupInquisitionAttempt records one Blue Knight's once-per-game Inquisition.
type CoupInquisitionAttempt struct {
	InquisitorID string
	TargetID     string
	CurrentLife  int
	PenaltyLife  int
	ConfirmedBy  string
	Resolved     bool
	Success      bool
}

// EnsureCoupInquisitionState returns initialized Inquisition state for a room.
func EnsureCoupInquisitionState(room *Room) *CoupInquisitionState {
	if room.CoupInquisition == nil {
		room.CoupInquisition = &CoupInquisitionState{}
	}
	if room.CoupInquisition.Attempts == nil {
		room.CoupInquisition.Attempts = make(map[string]CoupInquisitionAttempt)
	}
	return room.CoupInquisition
}

// CoupInquisitionPenalty returns half current life, rounded up.
func CoupInquisitionPenalty(currentLife int) int {
	if currentLife <= 0 {
		return 0
	}
	return (currentLife + 1) / 2
}

// NormalizeCoupInquisitionResultPolicy returns the default public result policy.
func NormalizeCoupInquisitionResultPolicy(policy CoupInquisitionResultPolicy) CoupInquisitionResultPolicy {
	switch policy {
	case CoupInquisitionResultPrivate:
		return CoupInquisitionResultPrivate
	default:
		return CoupInquisitionResultPublic
	}
}
