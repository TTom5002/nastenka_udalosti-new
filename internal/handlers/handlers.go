package handlers

import (
	"log"
	"nastenka_udalosti/internal/config"
	"nastenka_udalosti/internal/driver"
	"nastenka_udalosti/internal/forms"
	"nastenka_udalosti/internal/helpers"
	"nastenka_udalosti/internal/models"
	"nastenka_udalosti/internal/render"
	"nastenka_udalosti/internal/repository"
	"nastenka_udalosti/internal/repository/dbrepo"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

// TODO: Změň komentáře

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// TODO: Odkomentuj až budeš testovat
// // NewTestRepo creates a new repository
// func NewTestRepo(a *config.AppConfig) *Repository {
// 	return &Repository{
// 		App: a,
// 		DB:  dbrepo.NewTestingRepo(a),
// 	}
// }

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	events, err := m.DB.ShowEvents()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["events"] = events

	render.Template(w, r, "home.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// MakeEvent ukáže formulář pro vytvoření příspěvku - HOTOVO
func (m *Repository) MakeEvent(w http.ResponseWriter, r *http.Request) {

	var emptyEvent models.Event
	data := make(map[string]interface{})
	data["event"] = emptyEvent

	render.Template(w, r, "make-event.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) PostMakeEvent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Nelze parsovat formulář!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userInfo, err := helpers.GetUserInfo(r)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	event := models.Event{
		Header:   r.Form.Get("header"),
		Body:     r.Form.Get("body"),
		AuthorID: userInfo.ID,
	}

	form := forms.New(r.PostForm)

	form.Required("header", "body")
	// TODO: Délka nemusí být ale možná se bude hodit
	// form.MinLength("header", 3)

	if !form.Valid() {
		data := make(map[string]interface{})
		data["event"] = event
		render.Template(w, r, "make-event.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	err = m.DB.InsertEvent(event)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Nelze vytvořit příspěvek!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Úspěšně vytvořen příspěvek")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Login načte formulář pro přihlášení uživatele
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin příhlásí uživatele a získá od něm potřebné informace
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	user, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Chybné přihlašování údaje")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	userInfo := models.User{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		AccessLevel: user.AccessLevel,
		Verified:    user.Verified,
	}

	m.App.Session.Put(r.Context(), "userInfo", userInfo)
	m.App.Session.Put(r.Context(), "flash", "Úspěšně přihlášen")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout odhlásí uživatele
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	m.App.Session.Put(r.Context(), "flash", "Odhlášen")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) MyEvents(w http.ResponseWriter, r *http.Request) {
	userInfo, err := helpers.GetUserInfo(r)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	events, err := m.DB.ShowUserEvents(userInfo.ID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["events"] = events

	render.Template(w, r, "my-events.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) Signup(w http.ResponseWriter, r *http.Request) {
	var user models.User
	data := make(map[string]interface{})
	data["usersignup"] = user

	render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostLogin přihlásí uživatele
func (m *Repository) PostSignup(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	password := r.Form.Get("password")
	passwordver := r.Form.Get("passwordver")

	form := forms.New(r.PostForm)
	form.Required("firstname", "lastname", "email", "password", "passwordver")
	form.IsEmail("email")

	if !form.SamePassword(password, passwordver) {
		form.Errors.Add("password", "Hesla se neshodují")
		form.Errors.Add("passwordver", "Hesla se neshodují")
	}

	user := models.User{
		FirstName: r.Form.Get("firstname"),
		LastName:  r.Form.Get("lastname"),
		Email:     r.Form.Get("email"),
		Password:  r.Form.Get("password"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// TODO: Dodělej aby se data při chybném poslání formuláře nemazaly
	data := make(map[string]interface{})
	data["usersignup"] = user

	if !form.Valid() {
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// TODO: Do database nahraj data
	err = m.DB.SignUpUser(user)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Registrace se nepovedla")
		http.Redirect(w, r, "/user/singup/", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Úspěšně zaregistrován, vyčkejte na ověření")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) EditEvent(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	event_id, err := strconv.Atoi(exploded[5])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	event, err := m.DB.GetEventByID(event_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	userInfo, _ := helpers.GetUserInfo(r)
	if userInfo.ID != event.AuthorID {
		m.App.Session.Put(r.Context(), "flash", "Nejste autorem tohodle příspěvku")
		http.Redirect(w, r, "/dashboard/cu/posts/my-events", http.StatusSeeOther)
	}

	data := make(map[string]interface{})
	data["event"] = event

	// FIXME
	render.Template(w, r, "show-event.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) PostUpdateEvent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[5])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	event := models.Event{
		Header: r.Form.Get("header"),
		Body:   r.Form.Get("body"),
		ID:     id,
	}

	data := make(map[string]interface{})
	data["event"] = event

	form := forms.New(r.PostForm)
	form.Required("header", "body")
	if !form.Valid() {
		render.Template(w, r, "show-event.page.tmpl", &models.TemplateData{
			Data: data,
			Form: form,
		})
		return
	}

	err = m.DB.UpdateEventByID(event)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Příspěvek se upravil")
	http.Redirect(w, r, "/dashboard/cu/posts/my-events", http.StatusSeeOther)
}

func (m *Repository) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	event_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	event, err := m.DB.GetEventByID(event_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	userInfo, err := helpers.GetUserInfo(r)
	if err != nil {
		return
	}
	if userInfo.ID != event.AuthorID {
		m.App.Session.Put(r.Context(), "error", "Nejste autorem tohodle příspěvku")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		err = m.DB.DeleteEventByID(event_id)
		if err != nil {
			return
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Událost byla smazána")
	http.Redirect(w, r, "/dashboard/cu/posts/my-events", http.StatusSeeOther)
}

func (m *Repository) ShowAllUnverifiedUsers(w http.ResponseWriter, r *http.Request) {
	users, err := m.DB.ShowUnverifiedUsers()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["users"] = users

	render.Template(w, r, "admin-unverifiedusers.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) EditProfile(w http.ResponseWriter, r *http.Request) {
	userInfo, err := helpers.GetUserInfo(r)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["userInfo"] = userInfo

	render.Template(w, r, "profile.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) PostEditProfile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	userInfo, err := helpers.GetUserInfo(r)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	user := models.User{
		ID:        userInfo.ID,
		FirstName: r.Form.Get("firstname"),
		LastName:  r.Form.Get("lastname"),
		Email:     r.Form.Get("email"),
		Password:  r.Form.Get("password"),
	}

	data := make(map[string]interface{})
	data["userInfo"] = user

	form := forms.New(r.PostForm)
	form.Required("firstname", "lastname")

	if !form.SamePassword(r.Form.Get("password"), r.Form.Get("passwordver")) {
		form.Errors.Add("password", "Hesla se neshodují")
		form.Errors.Add("passwordver", "Hesla se neshodují")
	}

	if !form.Valid() {
		render.Template(w, r, "profile.page.tmpl", &models.TemplateData{
			Data: data,
			Form: form,
		})
		return
	}

	err = m.DB.UpdateProfile(user)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	userInfo, err = helpers.GetUserInfo(r)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	newUserInfo := models.User{
		ID:          userInfo.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       userInfo.Email,
		AccessLevel: userInfo.AccessLevel,
		Verified:    userInfo.Verified,
	}

	m.App.Session.Put(r.Context(), "userInfo", newUserInfo)
	m.App.Session.Put(r.Context(), "flash", "Profil se upravil")
	http.Redirect(w, r, "/dashboard/cu/profile", http.StatusSeeOther)
}

func (m *Repository) PostVerUsers(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	for name := range r.PostForm {
		if strings.HasPrefix(name, "to_ver") {
			exploded := strings.Split(name, "_")
			userID, _ := strconv.Atoi(exploded[2])
			err := m.DB.VerUserByID(userID)
			if err != nil {
				log.Println(err)
			}
		} else if strings.HasPrefix(name, "to_dec") {
			exploded := strings.Split(name, "_")
			userID, _ := strconv.Atoi(exploded[2])
			err := m.DB.DeleteUserByID(userID)
			if err != nil {
				log.Println(err)
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/dashboard/admin/unverified-users", http.StatusSeeOther)
}
