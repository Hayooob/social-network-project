package models

import "time"

type User struct {
	ID           int64      `db:"id"`
	UUID         string     `db:"uuid"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	FullName     string     `db:"full_name"`
	DateOfBirth  string     `db:"date_of_birth"`
	AvatarURL    *string    `db:"avatar_url"`
	Nickname     *string    `db:"nickname"`
	AboutMe      *string    `db:"about_me"`
	IsPrivate    bool       `db:"is_private"`
	CreatedAt    time.Time  `db:"created_at"`
}

//messages

//posts

