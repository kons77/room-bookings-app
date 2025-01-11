package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kons77/room-bookings-app/internal/config"
	"github.com/kons77/room-bookings-app/internal/driver"
	"github.com/kons77/room-bookings-app/internal/handlers"
	"github.com/kons77/room-bookings-app/internal/helpers"
	"github.com/kons77/room-bookings-app/internal/models"
	"github.com/kons77/room-bookings-app/internal/render"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close() // connection won't be closed until the main function itself stops running

	fmt.Printf("Starting application on port %s \n", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {

	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	//change this ti true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New() // := instead of = creates `variable shadowing` because session is declare out of func
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true                  // save after closing browser
	session.Cookie.SameSite = http.SameSiteLaxMode // how strict
	session.Cookie.Secure = app.InProduction       // insist the cookie to be encrypted - set true in production, for https

	app.Session = session

	// connect to db
	log.Println("Connecting to database... ")

	//delete entering password later
	var pswd string
	fmt.Println("Enter postgres db password (my standart digits pass without letters)")
	fmt.Scan(&pswd)

	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=" + pswd)
	if err != nil {
		log.Fatal("Cannot connet to database! Dying...")
	}
	log.Println("Connected to database")
	// defer db.SQL.Close() - it must be not here but in main function

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}
	app.TemplateCache = tc
	app.UseCache = app.InProduction

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
