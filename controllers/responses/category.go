package responses

type CreateCategory struct {
	Uuid string `json:"uuid"`
}

type Category struct {
	UUID   string  `json:"uuid"`
	Name   string  `json:"name"`
	Points float32 `json:"points,omitempty"`
}
