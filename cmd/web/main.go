package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"nastenka_udalosti/internal/config"
	"nastenka_udalosti/internal/driver"
	"nastenka_udalosti/internal/handlers"
	"nastenka_udalosti/internal/helpers"
	"nastenka_udalosti/internal/models"
	"nastenka_udalosti/internal/render"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// Main je hlavní funkce aplikace
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	fmt.Printf(fmt.Sprintf("Starting application on port %s", portNumber))
	fmt.Println()

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// TODO: Co budu vkládat do relace
	gob.Register(models.Event{})
	gob.Register(models.User{})
	gob.Register(map[string]int{})

	// TODO: Změň na true až půjde do produkce
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Přípojení k databázi
	log.Println("Connecting to database...")
	// TODO: Uprav název databáze
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=eventboarddb user=postgres password=tom")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to database!")
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
