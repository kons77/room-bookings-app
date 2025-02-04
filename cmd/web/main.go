package main

import (
	"encoding/gob"
	"flag"
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
	"gopkg.in/yaml.v3"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// DBCfgAlt holds database.yml - temp type until db yml cfg move to app.Config
type DBCfgAlt struct {
	Development struct {
		Dialect  string `yaml:"dialect"`
		Database string `yaml:"database"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Pool     int    `yaml:"pool"`
	} `yaml:"development"`
	/*
		Test struct {
			URL string `yaml:"url"`
		} `yaml:"test"`
		Production struct {
			URL string `yaml:"url"`
		} `yaml:"production"`
	*/
}

// getPasswordFromYaml reads password from database.yml - temp fucn until db yml cfg move to app.Config
func getPasswordFromYaml() (string, error) {
	data, err := os.ReadFile("database.yml")
	if err != nil {
		return "", err
	}

	var cfg DBCfgAlt
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Development.Password, nil
}

// main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close() // connection won't be closed until the main function itself stops running
	defer close(app.MailChan)

	fmt.Println("Starting mail listener...")
	listenForMail()

	fmt.Printf("Starting application on port %s \n", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {

	// Register custom types for session storage // what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	// read this from yaml file by func getPasswordFromYaml() to struct DBCfgAlt
	dbHost := flag.String("dbhost", "127.0.0.1", "Database host") // localhost
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPswd := flag.String("dbpswd", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database SSL settings (disable, prefer, require)")

	flag.Parse()

	_ = *dbPswd //

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	//
	app.InProduction = *inProduction
	app.UseCache = *useCache

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New() // := instead of = creates `variable shadowing` because session is declare out of func
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true                  // save after closing browser
	session.Cookie.SameSite = http.SameSiteLaxMode // how strict
	// Secure - insist the cookie to be encrypted - set true in production, for https
	session.Cookie.Secure = false //app.InProduction

	app.Session = session

	// connect to db
	log.Println("Connecting to database... ")

	// ! rewrite to Load defaults from yaml (test and production) - but override with flags if provided
	pswd, err := getPasswordFromYaml()
	// log.Println(pswd, err)
	if err != nil {
		log.Fatal(err)
	}

	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, pswd, *dbSSL)

	// ! rewrite to Load defaults from yaml (test and production) - but override with flags if provided
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connet to database! Dying...")
	}
	// defer db.SQL.Close() - it must be not here but in main function
	log.Println("Connected to database")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}
	app.TemplateCache = tc

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}

/* send_email_standard sends email using standart go library package - we do not use it
func send_email_standart() {
	from := "me@here.com"
	pswd := ""
	mailserver := "localhost"
	where := []string{
		"you@there.com",
	}
	msgContent := []byte("Hello, world")

	// credentials of mailserver
	auth := smtp.PlainAuth("", from, pswd, mailserver)
	err := smtp.SendMail("localhost:1025", auth, from, where, msgContent)
	if err != nil {
		log.Println("somethings goes wrong", err)
	}
}*/
