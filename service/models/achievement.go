package models

type SimpleAchievement struct {
	UUID           string  `json:"uuid"`
	AttachmentLink string  `json:"attachment_link"`
	Status         string  `json:"status"`
	CategoryName   string  `json:"category_name"`
	CategoryUUID   string  `json:"category_uuid"`
	AttachmentDate string  `json:"attachment_date,omitempty"`
	PointAmount    float32 `json:"point_amount"`
	Comment        string  `json:"comment"`
}

type Achievement struct {
	UUID           string
	AttachmentLink string
	Status         string
	Categories     []Category
	Comment        string
	PointAmount    float32
}
