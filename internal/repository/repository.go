package repository

import "nastenka_udalosti/internal/models"

// DatabaseRepo uchovává postgres funkce
type DatabaseRepo interface {
	AllUsers() bool
	InsertEvent(event models.Event) error
	// Authenticate(email, testPassword string) (int, string, bool, int, error)
	Authenticate(email, testPassword string) (models.User, error)
	ShowEvents() ([]models.Event, error)
	GetEventByID(event_id int) (models.Event, error)
	ShowUserEvents(id int) ([]models.Event, error)
	SignUpUser(user models.User) error
	DeleteEventByID(eventID int) error
	UpdateEventByID(event models.Event) error
	ShowUnverifiedUsers() ([]models.User, error)
	UpdateProfile(user models.User) error
}
