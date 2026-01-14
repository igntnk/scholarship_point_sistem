package requests

type CreateCategory struct {
	Name        string  `json:"name"`
	Comment     string  `json:"comment"`
	ParentUuid  string  `json:"parent_uuid"`
	PointAmount float32 `json:"point_amount"`
}
