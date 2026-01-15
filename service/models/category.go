package models

type Category struct {
	UUID               string  `json:"uuid"`
	Name               string  `json:"name"`
	ParentCategoryUUID string  `json:"parent_uuid"`
	PointAmount        float32 `json:"point_amount"`
	Comment            string  `json:"comment"`
	Status             string  `json:"status"`
}
