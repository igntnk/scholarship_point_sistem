package responses

type Achievement struct {
	UUID           string     `json:"uuid"`
	Comment        string     `json:"comment"`
	Status         string     `json:"status"`
	Categories     []Category `json:"categories"`
	PointAmount    float32    `json:"point_amount"`
	AttachmentLink string     `json:"attachment_link"`
}

type SimpleAchievement struct {
	UUID           string  `json:"uuid"`
	Comment        string  `json:"comment"`
	Status         string  `json:"status"`
	CategoryName   string  `json:"category_name"`
	PointAmount    float32 `json:"point_amount"`
	AttachmentLink string  `json:"attachment_link"`
}
