package models

import "time"

// TODO: Dodělej modely tabulek databáze

type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	Verified    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Event struct {
	ID        int
	Header    string
	Body      string
	AuthorID  int
	CreatedAt time.Time
	UpdatedAt time.Time
	User      User
}
