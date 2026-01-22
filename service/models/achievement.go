package models

type SimpleAchievement struct {
	UUID           string
	AttachmentLink string
	Status         string
	Comment        string
	CategoryName   string
	PointAmount    float32
}

type Achievement struct {
	UUID           string
	AttachmentLink string
	Status         string
	Categories     []Category
	Comment        string
	PointAmount    float32
}
