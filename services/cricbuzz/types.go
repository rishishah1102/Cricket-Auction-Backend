package cricbuzz

// ── Recent Matches Response ──────────────────────────────────────────────────

type RecentMatchesResponse struct {
	TypeMatches []TypeMatch `json:"typeMatches"`
}

type TypeMatch struct {
	MatchType     string        `json:"matchType"`
	SeriesMatches []SeriesMatch `json:"seriesMatches"`
}

type SeriesMatch struct {
	SeriesAdWrapper *SeriesAdWrapper `json:"seriesAdWrapper"`
}

type SeriesAdWrapper struct {
	SeriesID   int         `json:"seriesId"`
	SeriesName string      `json:"seriesName"`
	Matches    []MatchItem `json:"matches"`
}

type MatchItem struct {
	MatchInfo  MatchDetail `json:"matchInfo"`
	MatchScore *MatchScore `json:"matchScore"`
}

type MatchDetail struct {
	MatchID         int      `json:"matchId"`
	SeriesID        int      `json:"seriesId"`
	SeriesName      string   `json:"seriesName"`
	MatchDesc       string   `json:"matchDesc"`
	MatchFormat     string   `json:"matchFormat"`
	State           string   `json:"state"`
	Status          string   `json:"status"`
	Team1           TeamInfo `json:"team1"`
	Team2           TeamInfo `json:"team2"`
	StateTitle      string   `json:"stateTitle"`
	IsTimeAnnounced bool     `json:"isTimeAnnounced"`
}

type TeamInfo struct {
	TeamID   int    `json:"teamId"`
	TeamName string `json:"teamName"`
	TeamSName string `json:"teamSName"`
}

type MatchScore struct {
	Team1Score *TeamScore `json:"team1Score"`
	Team2Score *TeamScore `json:"team2Score"`
}

type TeamScore struct {
	Inngs1 *InningsScore `json:"inngs1"`
}

type InningsScore struct {
	InningsID int     `json:"inningsId"`
	Runs      int     `json:"runs"`
	Wickets   int     `json:"wickets"`
	Overs     float64 `json:"overs"`
}

// ── Scorecard Response ───────────────────────────────────────────────────────

type ScorecardResponse struct {
	Scorecard       []ScorecardInnings `json:"scorecard"`
	IsMatchComplete bool               `json:"ismatchcomplete"`
	Status          string             `json:"status"`
}

type ScorecardInnings struct {
	InningsID   int       `json:"inningsid"`
	Batsmen     []Batsman `json:"batsman"`
	Bowlers     []Bowler  `json:"bowler"`
	Score       int       `json:"score"`
	Wickets     int       `json:"wickets"`
	Overs       float64   `json:"overs"`
	BatTeamName string    `json:"batteamname"`
	BatTeamSName string   `json:"batteamsname"`
}

type Batsman struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Runs            int    `json:"runs"`
	Balls           int    `json:"balls"`
	Fours           int    `json:"fours"`
	Sixes           int    `json:"sixes"`
	StrikeRate      string `json:"strkrate"`
	OutDesc         string `json:"outdec"`
	IsCaptain       bool   `json:"iscaptain"`
	IsKeeper        bool   `json:"iskeeper"`
	InMatchChange   string `json:"inmatchchange"`
	IsOverseas      bool   `json:"isoverseas"`
}

type Bowler struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Overs         string `json:"overs"`
	Maidens       int    `json:"maidens"`
	Wickets       int    `json:"wickets"`
	Runs          int    `json:"runs"`
	Economy       string `json:"economy"`
	Dots          int    `json:"dots"`
	Balls         int    `json:"balls"`
	InMatchChange string `json:"inmatchchange"`
	IsOverseas    bool   `json:"isoverseas"`
}

// ── Simplified Match (returned by our API) ───────────────────────────────────

type SimplifiedMatch struct {
	MatchID    int    `json:"match_id"`
	MatchDesc  string `json:"match_desc"`
	Team1SName string `json:"team1_sname"`
	Team2SName string `json:"team2_sname"`
	Team1Name  string `json:"team1_name"`
	Team2Name  string `json:"team2_name"`
	Status     string `json:"status"`
	Team1Score string `json:"team1_score"`
	Team2Score string `json:"team2_score"`
}
