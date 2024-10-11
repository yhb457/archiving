package database

import "strings"

type User struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	ProfileImage string `json:"profile_image"`
	WeeklyGoal   string `json:"weekly_goal"`
	Email        string `json:"email"`
	Nickname     string `json:"nickname"`
}

func CreateUser(username, password, email, nickname string) error {
	_, err := db.Exec(`INSERT INTO users (username, password, email, nickname) VALUES ('` + username + `', '` + password + `', '` + email + `', '` + nickname + `')`)
	if err != nil {
		return err
	}

	return nil
}

func GetUserID(username string) (int, error) {
	uid := 0
	r := db.QueryRow(`SELECT id FROM users WHERE username='` + username + `'`)
	if err := r.Scan(&uid); err != nil {
		return uid, err
	}
	return uid, nil
}

func GetPassword(username string) (string, error) {
	pw := ""
	r := db.QueryRow(`SELECT password FROM users WHERE username='` + username + `'`)
	if err := r.Scan(&pw); err != nil {
		return pw, err
	}
	return pw, nil
}

func GetUser(userId string) (*User, error) {
	user := User{}
	r := db.QueryRow(`SELECT username, email, nickname, profile_image, weekly_goal FROM users WHERE id='` + userId + `'`)
	if err := r.Scan(&user.Username, &user.Email, &user.Nickname, &user.ProfileImage, &user.WeeklyGoal); err != nil {
		return &user, err
	}
	return &user, nil
}

func PutUser(userId, nickname, profileImage, weeklyGoal string) error {
	query := []string{}
	if nickname != "" {
		query = append(query, `nickname = '`+nickname+`'`)
	}
	if profileImage != "" {
		query = append(query, `profile_image = '`+profileImage+`'`)
	}
	if weeklyGoal != "" {
		query = append(query, `weekly_goal = '`+weeklyGoal+`'`)
	}
	c := strings.Join(query, ", ")

	_, err := db.Exec(`UPDATE users SET ` + c + ` WHERE id='` + userId + `'`)
	if err != nil {
		return err
	}

	return nil
}
