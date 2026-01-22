package requests

type UpsertAchievement struct {
	UUID           string   `json:"uuid"`
	AttachmentLink string   `json:"attachment_link"`
	Comment        string   `json:"comment"`
	CategoryUUIDs  []string `json:"category_uuids"`
}
