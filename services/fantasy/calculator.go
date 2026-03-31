package fantasy

import (
	"cric-auction-monolith/services/cricbuzz"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DismissalType represents how a batsman got out.
type DismissalType int

const (
	NotOut DismissalType = iota
	Caught
	Bowled
	LBW
	Stumped
	RunOut
	CaughtAndBowled
	RetiredHurt
	HitWicket
	Other
)

// DismissalInfo holds parsed information from an outdec string.
type DismissalInfo struct {
	Type          DismissalType
	Fielder       string // catcher or stumper
	Bowler        string
	RunOutThrower string // first name in run out (direct hit)
	RunOutAssist  string // second name in run out (not direct hit)
	IsDirectHit   bool
	RawText       string
}

// PointBreakdown shows how fantasy points were calculated.
type PointBreakdown struct {
	Batting  int      `json:"batting"`
	Bowling  int      `json:"bowling"`
	Fielding int      `json:"fielding"`
	Bonus    int      `json:"bonus"`
	Details  []string `json:"details"`
}

// PlayerPoints is the result for a single player.
type PlayerPoints struct {
	CricbuzzName string         `json:"cricbuzz_name"`
	Points       int            `json:"points"`
	Breakdown    PointBreakdown `json:"breakdown"`
}

var (
	reCaught       = regexp.MustCompile(`^c (.+?) b (.+)$`)
	reCaughtBowled = regexp.MustCompile(`^c and b (.+)$`)
	reStumped      = regexp.MustCompile(`^st (.+?) b (.+)$`)
	reRunOut2      = regexp.MustCompile(`^run out \((.+?)/(.+?)\)$`)
	reRunOut1      = regexp.MustCompile(`^run out \((.+?)\)$`)
	reLBW          = regexp.MustCompile(`^lbw b (.+)$`)
	reBowled       = regexp.MustCompile(`^b (.+)$`)
	reHitWicket    = regexp.MustCompile(`^hit wicket b (.+)$`)
)

// ParseDismissal extracts fielding and bowling info from an outdec string.
func ParseDismissal(outDec string) DismissalInfo {
	outDec = strings.TrimSpace(outDec)
	info := DismissalInfo{RawText: outDec}

	if outDec == "" || outDec == "not out" || strings.HasPrefix(outDec, "retired") {
		info.Type = NotOut
		if strings.HasPrefix(outDec, "retired") {
			info.Type = RetiredHurt
		}
		return info
	}

	if m := reHitWicket.FindStringSubmatch(outDec); m != nil {
		info.Type = HitWicket
		info.Bowler = strings.TrimSpace(m[1])
		return info
	}
	if m := reCaughtBowled.FindStringSubmatch(outDec); m != nil {
		info.Type = CaughtAndBowled
		info.Fielder = strings.TrimSpace(m[1])
		info.Bowler = strings.TrimSpace(m[1])
		return info
	}
	if m := reCaught.FindStringSubmatch(outDec); m != nil {
		info.Type = Caught
		info.Fielder = strings.TrimSpace(m[1])
		info.Bowler = strings.TrimSpace(m[2])
		return info
	}
	if m := reStumped.FindStringSubmatch(outDec); m != nil {
		info.Type = Stumped
		info.Fielder = strings.TrimSpace(m[1])
		info.Bowler = strings.TrimSpace(m[2])
		return info
	}
	if m := reRunOut2.FindStringSubmatch(outDec); m != nil {
		info.Type = RunOut
		info.RunOutThrower = strings.TrimSpace(m[1])
		info.RunOutAssist = strings.TrimSpace(m[2])
		info.IsDirectHit = false
		return info
	}
	if m := reRunOut1.FindStringSubmatch(outDec); m != nil {
		info.Type = RunOut
		info.RunOutThrower = strings.TrimSpace(m[1])
		info.IsDirectHit = true
		return info
	}
	if m := reLBW.FindStringSubmatch(outDec); m != nil {
		info.Type = LBW
		info.Bowler = strings.TrimSpace(m[1])
		return info
	}
	if m := reBowled.FindStringSubmatch(outDec); m != nil {
		info.Type = Bowled
		info.Bowler = strings.TrimSpace(m[1])
		return info
	}

	info.Type = Other
	return info
}

// CalculateAllPoints computes fantasy points for every player in a scorecard.
func CalculateAllPoints(scorecard *cricbuzz.ScorecardResponse) map[string]*PlayerPoints {
	results := make(map[string]*PlayerPoints)

	// Build a name alias map: maps all known name variants to a canonical name.
	// This fixes issues like "Philip Salt" (batsman name) vs "Phil Salt" (in outdec).
	aliases := buildNameAliases(scorecard)

	// Collect all dismissals from both innings (for fielding points).
	var allDismissals []DismissalInfo
	for _, inn := range scorecard.Scorecard {
		for _, bat := range inn.Batsmen {
			d := ParseDismissal(bat.OutDesc)
			allDismissals = append(allDismissals, d)
		}
	}

	// Track which players appeared (for dedup).
	type playerKey struct {
		name string
		team string
	}
	seen := make(map[playerKey]bool)

	// Process each innings.
	for _, inn := range scorecard.Scorecard {
		teamSName := inn.BatTeamSName

		// ── Batting points ──
		for _, bat := range inn.Batsmen {
			canonical := resolveAlias(aliases, bat.Name)
			key := playerKey{name: strings.ToLower(canonical), team: teamSName}
			if seen[key] {
				continue
			}
			seen[key] = true

			pp := getOrCreate(results, canonical)
			batting, details := calcBatting(bat)
			pp.Breakdown.Batting += batting
			pp.Breakdown.Details = append(pp.Breakdown.Details, details...)

			// +4 for playing in the match (every player who appeared).
			if pp.Breakdown.Bonus == 0 {
				pp.Breakdown.Bonus += 4
				pp.Breakdown.Details = append(pp.Breakdown.Details, "Playing in match (+4)")
			}
		}

		// ── Bowling points ──
		for _, bowl := range inn.Bowlers {
			canonical := resolveAlias(aliases, bowl.Name)
			pp := getOrCreate(results, canonical)
			bowling, details := calcBowling(bowl, inn.Batsmen)
			pp.Breakdown.Bowling += bowling
			pp.Breakdown.Details = append(pp.Breakdown.Details, details...)

			// +4 for playing (if not already given from batting).
			if pp.Breakdown.Bonus == 0 {
				pp.Breakdown.Bonus += 4
				pp.Breakdown.Details = append(pp.Breakdown.Details, "Playing in match (+4)")
			}
		}
	}

	// ── Fielding points (across entire match) ──
	fieldingStats := calcFieldingStats(allDismissals)
	for name, stats := range fieldingStats {
		canonical := resolveAlias(aliases, name)
		pp := getOrCreate(results, canonical)
		pp.Breakdown.Fielding += stats.points
		pp.Breakdown.Details = append(pp.Breakdown.Details, stats.details...)
	}

	// Sum totals.
	for _, pp := range results {
		pp.Points = pp.Breakdown.Batting + pp.Breakdown.Bowling + pp.Breakdown.Fielding + pp.Breakdown.Bonus
	}

	return results
}

// buildNameAliases creates a mapping from all name variants to a canonical name.
// It cross-references batsman names with names appearing in outdec strings.
// E.g., batsman "Philip Salt" and outdec "c Phil Salt b ..." → both map to "Philip Salt".
func buildNameAliases(scorecard *cricbuzz.ScorecardResponse) map[string]string {
	aliases := make(map[string]string) // lowercase variant → canonical name

	// Collect all canonical names (batsman + bowler names).
	var canonicalNames []string
	for _, inn := range scorecard.Scorecard {
		for _, bat := range inn.Batsmen {
			canonicalNames = append(canonicalNames, bat.Name)
			aliases[strings.ToLower(bat.Name)] = bat.Name
		}
		for _, bowl := range inn.Bowlers {
			if _, exists := aliases[strings.ToLower(bowl.Name)]; !exists {
				canonicalNames = append(canonicalNames, bowl.Name)
				aliases[strings.ToLower(bowl.Name)] = bowl.Name
			}
		}
	}

	// Collect all names from outdec strings.
	var outdecNames []string
	for _, inn := range scorecard.Scorecard {
		for _, bat := range inn.Batsmen {
			d := ParseDismissal(bat.OutDesc)
			if d.Fielder != "" {
				outdecNames = append(outdecNames, d.Fielder)
			}
			if d.Bowler != "" {
				outdecNames = append(outdecNames, d.Bowler)
			}
			if d.RunOutThrower != "" {
				outdecNames = append(outdecNames, d.RunOutThrower)
			}
			if d.RunOutAssist != "" {
				outdecNames = append(outdecNames, d.RunOutAssist)
			}
		}
	}

	// For each outdec name that doesn't have an exact match in canonical names,
	// try to find a match via last-name within the same set.
	for _, odName := range outdecNames {
		odLower := strings.ToLower(odName)
		if _, exists := aliases[odLower]; exists {
			continue // already mapped
		}
		// Try last-name match against canonical names.
		odLast := lastWord(odLower)
		if len(odLast) <= 2 {
			continue
		}
		for _, cn := range canonicalNames {
			cnLast := lastWord(strings.ToLower(cn))
			if strings.EqualFold(odLast, cnLast) {
				aliases[odLower] = cn
				break
			}
		}
		// If still not found, map to itself.
		if _, exists := aliases[odLower]; !exists {
			aliases[odLower] = odName
		}
	}

	return aliases
}

// resolveAlias returns the canonical name for a given name.
func resolveAlias(aliases map[string]string, name string) string {
	if canonical, ok := aliases[strings.ToLower(name)]; ok {
		return canonical
	}
	return name
}

func getOrCreate(m map[string]*PlayerPoints, name string) *PlayerPoints {
	// Use the canonical name (first occurrence).
	for k, v := range m {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	pp := &PlayerPoints{CricbuzzName: name}
	m[name] = pp
	return pp
}

// ── Batting ──────────────────────────────────────────────────────────────────

func calcBatting(bat cricbuzz.Batsman) (int, []string) {
	points := 0
	var details []string

	// Runs.
	if bat.Runs > 0 {
		points += bat.Runs
		details = append(details, fmt.Sprintf("%d runs (+%d)", bat.Runs, bat.Runs))
	}

	// Boundary bonus.
	if bat.Fours > 0 {
		bonus := bat.Fours * 4
		points += bonus
		details = append(details, fmt.Sprintf("%dx4s (+%d)", bat.Fours, bonus))
	}

	// Six bonus.
	if bat.Sixes > 0 {
		bonus := bat.Sixes * 6
		points += bonus
		details = append(details, fmt.Sprintf("%dx6s (+%d)", bat.Sixes, bonus))
	}

	// Milestone bonus (highest applicable only).
	switch {
	case bat.Runs >= 100:
		points += 16
		details = append(details, "Century bonus (+16)")
	case bat.Runs >= 75:
		points += 12
		details = append(details, "75-run bonus (+12)")
	case bat.Runs >= 50:
		points += 8
		details = append(details, "Half-century bonus (+8)")
	case bat.Runs >= 25:
		points += 4
		details = append(details, "25-run bonus (+4)")
	}

	// Duck.
	isOut := bat.OutDesc != "" && bat.OutDesc != "not out" && !strings.HasPrefix(bat.OutDesc, "retired")
	if bat.Runs == 0 && isOut && bat.Balls > 0 {
		// Duck applies to BAT, WK, AR (controller will filter by DB role).
		points -= 2
		details = append(details, "Duck (-2)")
	}

	// Strike rate bonus/penalty (min 10 balls).
	if bat.Balls >= 10 {
		sr := (float64(bat.Runs) / float64(bat.Balls)) * 100
		srBonus := calcStrikeRateBonus(sr)
		if srBonus != 0 {
			points += srBonus
			details = append(details, fmt.Sprintf("SR %.1f (%+d)", sr, srBonus))
		}
	}

	return points, details
}

func calcStrikeRateBonus(sr float64) int {
	switch {
	case sr > 170:
		return 6
	case sr > 150:
		return 4
	case sr >= 130:
		return 2
	case sr >= 70.01 && sr < 130:
		return 0
	case sr >= 60:
		return -2
	case sr >= 50:
		return -4
	case sr < 50:
		return -6
	}
	return 0
}

// ── Bowling ──────────────────────────────────────────────────────────────────

func calcBowling(bowl cricbuzz.Bowler, batsmen []cricbuzz.Batsman) (int, []string) {
	points := 0
	var details []string

	// Count wickets and LBW/Bowled bonuses from dismissal strings.
	wickets := 0
	lbwBowledCount := 0
	for _, bat := range batsmen {
		d := ParseDismissal(bat.OutDesc)
		if d.Bowler == "" {
			continue
		}
		if !namesMatch(d.Bowler, bowl.Name) {
			continue
		}
		// Run outs don't count as bowler wickets.
		if d.Type == RunOut {
			continue
		}
		wickets++
		if d.Type == LBW || d.Type == Bowled {
			lbwBowledCount++
		}
	}

	// Wicket points.
	if wickets > 0 {
		wPts := wickets * 30
		points += wPts
		details = append(details, fmt.Sprintf("%d wickets (+%d)", wickets, wPts))
	}

	// LBW/Bowled bonus.
	if lbwBowledCount > 0 {
		bonus := lbwBowledCount * 8
		points += bonus
		details = append(details, fmt.Sprintf("%dx LBW/Bowled bonus (+%d)", lbwBowledCount, bonus))
	}

	// Wicket milestone bonus (highest applicable only).
	switch {
	case wickets >= 5:
		points += 12
		details = append(details, "5-wicket bonus (+12)")
	case wickets >= 4:
		points += 8
		details = append(details, "4-wicket bonus (+8)")
	case wickets >= 3:
		points += 4
		details = append(details, "3-wicket bonus (+4)")
	}

	// Maiden overs.
	if bowl.Maidens > 0 {
		bonus := bowl.Maidens * 12
		points += bonus
		details = append(details, fmt.Sprintf("%d maiden(s) (+%d)", bowl.Maidens, bonus))
	}

	// Economy rate bonus/penalty (min 2 overs).
	overs := parseOvers(bowl.Overs)
	if overs >= 2.0 {
		eco := parseFloat(bowl.Economy)
		ecoBonus := calcEconomyBonus(eco)
		if ecoBonus != 0 {
			points += ecoBonus
			details = append(details, fmt.Sprintf("Economy %.1f (%+d)", eco, ecoBonus))
		}
	}

	return points, details
}


// calcEconomyBonus returns economy rate bonus/penalty WITH dot ball compensation.
// Since Cricbuzz API doesn't provide dot ball data, the economy brackets
// include a dot-ball proxy: lower economy → more dots → higher bonus.
func calcEconomyBonus(eco float64) int {
	switch {
	case eco < 5:
		return 22 // +6 ER + 16 Dot Proxy
	case eco < 6:
		return 16 // +4 ER + 12 Dot Proxy
	case eco <= 7:
		return 10 // +2 ER + 10 Dot Proxy
	case eco > 7 && eco < 10:
		return 7 // 0 ER + 7 Dot Proxy
	case eco <= 11:
		return 0 // -2 ER + 2 Dot Proxy
	case eco <= 12:
		return -4
	case eco > 12:
		return -6
	}
	return 0
}

// ── Fielding ─────────────────────────────────────────────────────────────────

type fieldingResult struct {
	points  int
	details []string
}

func calcFieldingStats(dismissals []DismissalInfo) map[string]*fieldingResult {
	stats := make(map[string]*fieldingResult)

	getResult := func(name string) *fieldingResult {
		// Case-insensitive lookup.
		for k, v := range stats {
			if strings.EqualFold(k, name) {
				return v
			}
		}
		r := &fieldingResult{}
		stats[name] = r
		return r
	}

	catchCount := make(map[string]int) // lowercase name -> count

	for _, d := range dismissals {
		switch d.Type {
		case Caught, CaughtAndBowled:
			if d.Fielder != "" {
				r := getResult(d.Fielder)
				r.points += 8
				r.details = append(r.details, "Catch (+8)")
				catchCount[strings.ToLower(d.Fielder)]++
			}
		case Stumped:
			if d.Fielder != "" {
				r := getResult(d.Fielder)
				r.points += 12
				r.details = append(r.details, "Stumping (+12)")
			}
		case RunOut:
			if d.RunOutThrower != "" {
				if d.IsDirectHit {
					r := getResult(d.RunOutThrower)
					r.points += 12
					r.details = append(r.details, "Run out direct hit (+12)")
				} else {
					r := getResult(d.RunOutThrower)
					r.points += 6
					r.details = append(r.details, "Run out throw (+6)")
					if d.RunOutAssist != "" {
						r2 := getResult(d.RunOutAssist)
						r2.points += 6
						r2.details = append(r2.details, "Run out assist (+6)")
					}
				}
			}
		}
	}

	// 3-catch bonus.
	for name, count := range catchCount {
		if count >= 3 {
			r := getResult(name)
			r.points += 4
			r.details = append(r.details, fmt.Sprintf("%d catches bonus (+4)", count))
		}
	}

	return stats
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func parseOvers(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// namesMatch checks if two player names refer to the same person.
// Uses case-insensitive comparison and last-name fallback.
func namesMatch(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if strings.EqualFold(a, b) {
		return true
	}
	// Compare last names.
	aLast := lastWord(a)
	bLast := lastWord(b)
	return len(aLast) > 2 && strings.EqualFold(aLast, bLast)
}

func lastWord(s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
