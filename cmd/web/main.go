package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kons77/room-bookings-app/internal/config"
	"github.com/kons77/room-bookings-app/internal/handlers"
	"github.com/kons77/room-bookings-app/internal/models"
	"github.com/kons77/room-bookings-app/internal/render"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

// main is the main application function
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting application on port %s \n", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error {

	// what am I going to put in the session
	gob.Register(models.Reservation{})

	//change this ti true when in production
	app.InProduction = false

	session = scs.New() // := instead of = creates `variable shadowing` because session is declare out of func
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true                  // save after closing browser
	session.Cookie.SameSite = http.SameSiteLaxMode // how strict
	session.Cookie.Secure = app.InProduction       // insist the cookie to be encrypted - set true in production, for https

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return err
	}
	app.TemplateCache = tc
	app.UseCache = app.InProduction

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)

	return nil
}
