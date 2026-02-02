package requests

type CreateCategory struct {
	Name       string           `json:"name"`
	ParentUuid string           `json:"parent_uuid"`
	Points     float32          `json:"points"`
	Values     []CategoryValues `json:"values"`
}

type CategoryValues struct {
	Name   string  `json:"name"`
	Points float32 `json:"points"`
}

type UpdateCategory struct {
	UUID   string           `json:"uuid"`
	Name   string           `json:"name"`
	Points float32          `json:"points"`
	Values []CategoryValues `json:"values"`
}
