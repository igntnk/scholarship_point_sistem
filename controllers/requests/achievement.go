package requests

type UpsertAchievement struct {
	UUID            string        `json:"uuid"`
	AttachmentLink  string        `json:"attachment_link"`
	Comment         string        `json:"comment"`
	CategoryUUID    string        `json:"category_uuid"`
	AchievementDate string        `json:"achievement_date"`
	Subcategories   []Subcategory `json:"subcategories"`
}

type Subcategory struct {
	UUID          string `json:"uuid"`
	SelectedValue string `json:"selected_value"`
}
