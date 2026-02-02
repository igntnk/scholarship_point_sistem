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
	CategoryUUID   string  `json:"category_uuid"`
	PointAmount    float32 `json:"point_amount"`
	AttachmentLink string  `json:"attachment_link"`
}

type FullAchievement struct {
	UUID            string        `json:"uuid"`
	Comment         string        `json:"comment"`
	AttachmentLink  string        `json:"attachment_link"`
	Status          string        `json:"status"`
	Category        Category      `json:"category"`
	AchievementDate string        `json:"achievement_date"`
	PointAmount     float32       `json:"point_amount"`
	Subcategories   []Subcategory `json:"subcategories"`
}
