package responses

type User struct {
	UUID                   string  `json:"uuid"`
	Name                   string  `json:"name"`
	SecondName             string  `json:"second_name"`
	Patronymic             string  `json:"patronymic"`
	BirthDate              string  `json:"birth_date"`
	PhoneNumber            string  `json:"phone_number"`
	GradebookNumber        string  `json:"gradebook_number"`
	Email                  string  `json:"email"`
	PointsAmount           float64 `json:"points_amount"`
	AchievementAmount      int     `json:"achievement_amount"`
	Valid                  bool    `json:"valid"`
	AllAchievementVerified bool    `json:"all_achievement_verified"`
}
