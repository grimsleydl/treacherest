package game

// CoupGreenHuntRequirement controls how many Blue Knights must die before
// King Fall to satisfy Green's Hunt.
type CoupGreenHuntRequirement string

const (
	CoupGreenHuntOneBlue  CoupGreenHuntRequirement = "one_blue"
	CoupGreenHuntAllBlues CoupGreenHuntRequirement = "all_blues"
)

// CoupInquisitionAmnesty controls whether successful Inquisition creates only
// King-side Green amnesty or also Broad Amnesty for Red-side sharing.
type CoupInquisitionAmnesty string

const (
	CoupInquisitionAmnestyKingSideOnly CoupInquisitionAmnesty = "king_side_only"
	CoupInquisitionAmnestyBroad        CoupInquisitionAmnesty = "broad"
)

func NormalizeCoupGreenHuntRequirement(requirement CoupGreenHuntRequirement) CoupGreenHuntRequirement {
	switch requirement {
	case CoupGreenHuntAllBlues:
		return CoupGreenHuntAllBlues
	default:
		return CoupGreenHuntOneBlue
	}
}

func NormalizeCoupInquisitionAmnesty(amnesty CoupInquisitionAmnesty) CoupInquisitionAmnesty {
	switch amnesty {
	case CoupInquisitionAmnestyBroad:
		return CoupInquisitionAmnestyBroad
	default:
		return CoupInquisitionAmnestyKingSideOnly
	}
}
