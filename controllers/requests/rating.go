package requests

type GetRating struct {
	SearchString string `json:"search"`
	Valid        bool   `json:"valid"`
	Winners      bool   `json:"winners"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}
