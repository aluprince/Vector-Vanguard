package scripts

//import "github.com/aluprince/Vector-Vanguard/scripts"



func CalculateLeadScore(audit AuditReport) int {
	score := 0

	// Pain Point 1: No Digital Presence (Highest Opportunity)
	if audit.NoWebsite == true{
		score += 50
	}
	// Pain Point 2: Broken Front Door
	if audit.IsBroken == true{
		score += 40
	}
	// Pain Point 3: The "Slow Giant" (Performance issues)
	if audit.LoadTimeSec > 5.0 {
		score += 25
	}
	// Pain Point 4: Abandoned Site
	if audit.OutdatedCopy {
		score += 15
	}

	return score
}