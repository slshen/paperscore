package game

type Modifiers []string

const (
	Throwing     = "TH"
	SacrificeHit = "SH"
	SacrificeFly = "SF"
	Obstruction  = "OBS"
)

type HitTrajectory string // BP, BG, BL, DP, F, FDP, G, GDP, GTP, IF, IPHR, L, LDP, LTP, P
type HitLocation string

func (mods Modifiers) Contains(codes ...string) bool {
	for _, m := range mods {
		for _, code := range codes {
			if m == code {
				return true
			}
		}
	}
	return false
}
