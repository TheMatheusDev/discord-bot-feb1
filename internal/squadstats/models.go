package squadstats

import "time"

// ServerPlayer represents the server and match data returned by /serversPlayers/{id}.json
type ServerPlayer struct {
	TeamID           string    `json:"teamID"`
	TeamName         string    `json:"teamName"`
	MatchID          int       `json:"matchID"`
	PublicSlots      int       `json:"publicSlots"`
	ReserveSlots     int       `json:"reserveSlots"`
	PublicQueue      int       `json:"publicQueue"`
	ReserveQueue     int       `json:"reserveQueue"`
	Time             time.Time `json:"time"`
	SimplePlayerList int       `json:"simplePlayerList"`
	// We omit `squads` and `players` for brevity since we only need matchID.
}

// MatchResponse represents the overall response from /matches?matchID={matchID}
type MatchResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Match   MatchDetails `json:"match"`
}

// MatchDetails contains the specific details about a match, like map and layer.
type MatchDetails struct {
	ID             int       `json:"id"`
	DLC            string    `json:"dlc"`
	Mod            string    `json:"mod"`
	Gamemode       string    `json:"gamemode"`
	MapClassname   string    `json:"mapClassname"`
	LayerClassname string    `json:"layerClassname"`
	Map            string    `json:"map"`
	Layer          string    `json:"layer"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	ServerID       int       `json:"serverID"`
}
