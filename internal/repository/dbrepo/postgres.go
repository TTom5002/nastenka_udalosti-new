package dbrepo

import (
	"context"
	"errors"
	"nastenka_udalosti/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// TODO: Udělej všechny databázové dotazy

// ShowEvents zobrazí určitý počet příspěvků
func (m *postgresDBRepo) ShowEvents() ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var events []models.Event

	// TODO: Kde je limit, tak budeš moct přidávat více příspěvků na stránku a offset jakou stránku
	query := `
		select e.event_id, e.event_header, e.event_body, e.event_created_at, e.event_updated_at,
		u.user_id, u.user_lastname
		from events e
		left join users u on (e.event_author_id = u.user_id)
		order by e.event_created_at desc
		LIMIT 25 offset 0
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return events, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Event
		err := rows.Scan(
			&i.ID,
			&i.Header,
			&i.Body,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.User.ID,
			&i.User.LastName,
		)

		if err != nil {
			return events, err
		}
		events = append(events, i)
	}

	if err = rows.Err(); err != nil {
		return events, err
	}

	return events, nil
}

func (m *postgresDBRepo) InsertEvent(event models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		insert into events(event_header, event_body, event_author_id, event_created_at, event_updated_at)
		values ($1,$2,$3,$4,$5)
	`
	_, err := m.DB.ExecContext(ctx, query,
		event.Header,
		event.Body,
		event.AuthorID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}
	return nil
}

// Authenticate ověří uživatele, že je přihlášen
func (m *postgresDBRepo) Authenticate(email, testPassword string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u models.User

	row := m.DB.QueryRowContext(ctx, "select user_id, user_firstname, user_lastname, user_email, user_password, user_verified, user_access_level from users where user_email = $1", email)
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.Verified,
		&u.AccessLevel,
	)
	if err != nil {
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return models.User{}, errors.New("chybné heslo")
	} else if err != nil {
		return models.User{}, err
	}
	u.Password = ""

	return u, nil
}

// Query až budu chtít vybírat podle role
/*
select e.event_id, e.event_header, e.event_body, e.event_created_at, e.event_updated_at,
		u.user_id, u.user_lastname
		from events e
		left join users u on (e.event_author_id = u.user_id)
		where u.user_access_level = 3
		order by e.event_created_at asc
		LIMIT 5 offset 0
*/

func (m *postgresDBRepo) ShowUserEvents(id int) ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var events []models.Event

	// TODO: Kde je limit, tak budeš moct přidávat více příspěvků na stránku a offset jakou stránku
	query := `
		select e.event_id, e.event_header, e.event_body, e.event_created_at, e.event_updated_at,
		u.user_id, u.user_lastname
		from events e
		left join users u on (e.event_author_id = u.user_id)
		where u.user_id  = $1
		order by e.event_created_at desc
		LIMIT 25 offset 0
	`

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return events, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Event
		err := rows.Scan(
			&i.ID,
			&i.Header,
			&i.Body,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.User.ID,
			&i.User.LastName,
		)

		if err != nil {
			return events, err
		}
		events = append(events, i)
	}

	if err = rows.Err(); err != nil {
		return events, err
	}

	return events, nil
}

func (m *postgresDBRepo) SignUpUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	user.Password = string(hashedPassword)

	query := `
		INSERT INTO users (user_firstname, user_lastname, user_email, user_password, user_created_at, user_updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`
	_, err := m.DB.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) GetEventByID(eventID int) (models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var event models.Event

	query := `
		SELECT e.event_id, e.event_header, e.event_body, e.event_author_id, e.event_created_at, e.event_updated_at, u.user_id, u.user_firstname, u.user_lastname  
		FROM events e
		left join users u on (e.event_author_id = u.user_id)
		where e.event_id = $1;
	`
	row := m.DB.QueryRowContext(ctx, query, eventID)
	err := row.Scan(
		&event.ID,
		&event.Header,
		&event.Body,
		&event.AuthorID,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.User.ID,
		&event.User.FirstName,
		&event.User.LastName,
	)

	if err != nil {
		return event, err
	}

	return event, nil
}

func (m *postgresDBRepo) UpdateEventByID(event models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update events set event_header = $1, event_body = $2, event_updated_at = $3 where event_id = $4
	`

	_, err := m.DB.ExecContext(ctx, query,
		event.Header,
		event.Body,
		time.Now(),
		event.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) DeleteEventByID(eventID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		delete from events where event_id = $1
	`

	_, err := m.DB.ExecContext(ctx, query, eventID)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Kde je limit, tak budeš moct přidávat více příspěvků na stránku a offset jakou stránku
// query := `
// select e.event_id, e.event_header, e.event_body, e.event_created_at, e.event_updated_at,
// u.user_id, u.user_lastname
// from events e
// left join users u on (e.event_author_id = u.user_id)
// where u.user_id  = $1
// order by e.event_created_at asc
// LIMIT 5 offset 0
// `

// VerIsAuthor zjistí jestli je uživatel autorem příspěvku
func (m *postgresDBRepo) VerIsAuthor(eventID, userID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT CASE WHEN COUNT(*) > 0 THEN true ELSE false END AS event_author_id
	FROM events
	WHERE event_id = $1 AND event_author_id = $2
	`

	var IsAuthor bool

	row := m.DB.QueryRowContext(ctx, query,
		eventID,
		userID,
	)
	err := row.Scan(
		&IsAuthor,
	)

	if err != nil {
		return false, err
	}

	if !IsAuthor {
		return false, err
	}

	return true, err
}

// ShowEvents zobrazí určitý počet příspěvků
func (m *postgresDBRepo) ShowUnverifiedUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var users []models.User

	// TODO: Kde je limit, tak budeš moct přidávat více příspěvků na stránku a offset jakou stránku
	query := `
		select user_id, user_firstname, user_lastname, user_email, user_updated_at
		from users
		where user_verified = false
		order by user_updated_at desc
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.User
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.UpdatedAt,
		)

		if err != nil {
			return users, err
		}
		users = append(users, i)
	}

	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}

func (m *postgresDBRepo) UpdateProfile(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set user_firstname = $1, user_lastname = $2, user_updated_at = $3 where user_id = $4
	`

	_, err := m.DB.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return err
	}

	if user.Password != "" {

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
		user.Password = string(hashedPassword)

		query := `
		update users set user_password = $1 where user_id = $2
		`

		_, err := m.DB.ExecContext(ctx, query,
			user.Password,
			user.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *postgresDBRepo) DeleteUserByID(userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	DELETE FROM users WHERE user_id=$1;
	`

	_, err := m.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

// VerUserByID změní ověření uživatele na true podle ID
func (m *postgresDBRepo) VerUserByID(userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	UPDATE users
	SET user_verified=true, user_updated_at=$1
	WHERE user_id=$2;	
	`

	_, err := m.DB.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return err
	}
	return nil
}
