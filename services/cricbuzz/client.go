package cricbuzz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey  string
	apiHost string
	http    *http.Client
}

func NewClient(apiKey, apiHost string) *Client {
	if apiHost == "" {
		apiHost = "crickbuzz-official-apis.p.rapidapi.com"
	}
	return &Client{
		apiKey:  apiKey,
		apiHost: apiHost,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) doRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", c.apiKey)
	req.Header.Set("x-rapidapi-host", c.apiHost)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	return body, nil
}

// FetchRecentMatches returns recent matches from Cricbuzz.
func (c *Client) FetchRecentMatches() (*RecentMatchesResponse, error) {
	url := fmt.Sprintf("https://%s/matches/recent", c.apiHost)
	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var result RecentMatchesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode recent matches: %w", err)
	}
	return &result, nil
}

// FetchScorecard returns the full scorecard for a given match ID.
func (c *Client) FetchScorecard(matchID int) (*ScorecardResponse, error) {
	url := fmt.Sprintf("https://%s/match/%d/scorecard", c.apiHost, matchID)
	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var result ScorecardResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode scorecard: %w", err)
	}
	return &result, nil
}

// FilterIPLMatches extracts completed IPL matches from the recent matches response.
func FilterIPLMatches(resp *RecentMatchesResponse) []SimplifiedMatch {
	var matches []SimplifiedMatch

	for _, tm := range resp.TypeMatches {
		if tm.MatchType != "League" {
			continue
		}
		for _, sm := range tm.SeriesMatches {
			if sm.SeriesAdWrapper == nil {
				continue
			}
			if !strings.Contains(strings.ToLower(sm.SeriesAdWrapper.SeriesName), "indian premier league") {
				continue
			}
			for _, m := range sm.SeriesAdWrapper.Matches {
				if m.MatchInfo.State != "Complete" {
					continue
				}
				simplified := SimplifiedMatch{
					MatchID:    m.MatchInfo.MatchID,
					MatchDesc:  m.MatchInfo.MatchDesc,
					Team1SName: m.MatchInfo.Team1.TeamSName,
					Team2SName: m.MatchInfo.Team2.TeamSName,
					Team1Name:  m.MatchInfo.Team1.TeamName,
					Team2Name:  m.MatchInfo.Team2.TeamName,
					Status:     m.MatchInfo.Status,
				}
				if m.MatchScore != nil {
					if m.MatchScore.Team1Score != nil && m.MatchScore.Team1Score.Inngs1 != nil {
						s := m.MatchScore.Team1Score.Inngs1
						simplified.Team1Score = fmt.Sprintf("%d/%d (%.1f)", s.Runs, s.Wickets, s.Overs)
					}
					if m.MatchScore.Team2Score != nil && m.MatchScore.Team2Score.Inngs1 != nil {
						s := m.MatchScore.Team2Score.Inngs1
						simplified.Team2Score = fmt.Sprintf("%d/%d (%.1f)", s.Runs, s.Wickets, s.Overs)
					}
				}
				matches = append(matches, simplified)
			}
		}
	}

	return matches
}
