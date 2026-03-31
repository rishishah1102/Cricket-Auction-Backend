package fantasy

import (
	"strings"
)

// DBPlayer represents a player from the auction database.
type DBPlayer struct {
	PlayerName string
	IPLTeam    string
	Role       string
	MatchID    string // MongoDB ObjectID as hex string
}

// MatchResult holds the result of matching a Cricbuzz player to a DB player.
type MatchResult struct {
	DBIndex    int     // index into the DB players slice (-1 if no match)
	Confidence float64 // 1.0 = exact, 0.8 = last-name, 0.6 = substring
}

// MatchPlayerToDB matches a Cricbuzz player name to the best DB player.
// cricbuzzTeam is the short team name from the scorecard innings (e.g., "CSK").
func MatchPlayerToDB(cricbuzzName string, cricbuzzTeam string, dbPlayers []DBPlayer) MatchResult {
	cbLower := strings.ToLower(strings.TrimSpace(cricbuzzName))

	// Pass 1: Exact match (case-insensitive).
	for i, p := range dbPlayers {
		if strings.EqualFold(strings.TrimSpace(p.PlayerName), cbLower) {
			return MatchResult{DBIndex: i, Confidence: 1.0}
		}
	}

	// Pass 2: Last-name match within same IPL team.
	cbLast := lastWord(cbLower)
	if len(cbLast) > 2 {
		for i, p := range dbPlayers {
			if !teamMatches(cricbuzzTeam, p.IPLTeam) {
				continue
			}
			dbLast := lastWord(strings.ToLower(p.PlayerName))
			if strings.EqualFold(cbLast, dbLast) {
				return MatchResult{DBIndex: i, Confidence: 0.8}
			}
		}
	}

	// Pass 3: Substring containment within same IPL team.
	for i, p := range dbPlayers {
		if !teamMatches(cricbuzzTeam, p.IPLTeam) {
			continue
		}
		pLower := strings.ToLower(strings.TrimSpace(p.PlayerName))
		if len(pLower) >= 3 && (strings.Contains(cbLower, pLower) || strings.Contains(pLower, cbLower)) {
			return MatchResult{DBIndex: i, Confidence: 0.6}
		}
	}

	return MatchResult{DBIndex: -1, Confidence: 0}
}

// teamMatches checks if two team identifiers refer to the same team.
// Handles both short names ("CSK") and full names ("Chennai Super Kings").
func teamMatches(a, b string) bool {
	a = strings.ToUpper(strings.TrimSpace(a))
	b = strings.ToUpper(strings.TrimSpace(b))
	if a == b {
		return true
	}

	// Map of short name -> possible full names / variations.
	teamAliases := map[string][]string{
		"CSK":  {"CHENNAI SUPER KINGS", "CHENNAI"},
		"MI":   {"MUMBAI INDIANS", "MUMBAI"},
		"RCB":  {"ROYAL CHALLENGERS BENGALURU", "ROYAL CHALLENGERS BANGALORE", "BENGALURU", "BANGALORE"},
		"RR":   {"RAJASTHAN ROYALS", "RAJASTHAN"},
		"KKR":  {"KOLKATA KNIGHT RIDERS", "KOLKATA"},
		"SRH":  {"SUNRISERS HYDERABAD", "SUNRISERS", "HYDERABAD"},
		"DC":   {"DELHI CAPITALS", "DELHI"},
		"PBKS": {"PUNJAB KINGS", "PUNJAB"},
		"GT":   {"GUJARAT TITANS", "GUJARAT"},
		"LSG":  {"LUCKNOW SUPER GIANTS", "LUCKNOW"},
	}

	resolveShort := func(name string) string {
		if _, ok := teamAliases[name]; ok {
			return name
		}
		for short, aliases := range teamAliases {
			for _, alias := range aliases {
				if name == alias {
					return short
				}
			}
		}
		return name
	}

	return resolveShort(a) == resolveShort(b)
}
