package entities

type User struct {
	ID       int64  `gorm:"primary_key;auto_increment" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"password"`
}
