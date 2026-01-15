package requests

type CreateCategory struct {
	Name        string  `json:"name"`
	Comment     string  `json:"comment"`
	ParentUuid  string  `json:"parent_uuid"`
	PointAmount float32 `json:"point_amount"`
}

type UpdateCategory struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Comment     string  `json:"comment"`
	ParentUuid  string  `json:"parent_uuid"`
	PointAmount float32 `json:"point_amount"`
	Status      string  `json:"status"`
}
