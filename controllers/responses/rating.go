package responses

type RatingShortInfo struct {
	CurrentPosition int     `json:"current_pos"`
	CurrentPoints   float64 `json:"current_points"`
	LeaderPoints    float64 `json:"leader_points"`
}
