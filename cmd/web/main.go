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
type yamlConfig struct {
	Development struct {
		Dialect  string `yaml:"dialect"`
		Database string `yaml:"database"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Pool     int    `yaml:"pool"`
	} `yaml:"development"`
}

// loadYamlConfig loads settings from database.yml
func loadYamlConfig() (yamlConfig, error) {
	var yConfig yamlConfig
	data, err := os.ReadFile("database.yml")
	if err != nil {
		return yConfig, err
	}

	err = yaml.Unmarshal(data, &yConfig)
	if err != nil {
		return yConfig, err
	}

	return yConfig, nil
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

	// register custom types for session storage (these types will be stored in the session)
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// load settings from the YAML configuration file
	dbConfig, err := loadYamlConfig()
	if err != nil {
		log.Println("Cannot read yaml file", err)
	}

	// read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "127.0.0.1", "Database host") // localhost
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPswd := flag.String("dbpswd", "", "Database password")
	dbPort := flag.String("dbport", "", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database SSL settings (disable, prefer, require)")

	flag.Parse()

	// Override YAML settings with flag values (if provided)
	if *dbHost != "127.0.0.1" {
		dbConfig.Development.Host = *dbHost
	}
	if *dbName != "" {
		dbConfig.Development.Database = *dbName
	}
	if *dbUser != "" {
		dbConfig.Development.User = *dbUser
	}
	if *dbPswd != "" {
		dbConfig.Development.Password = *dbPswd
	}
	if *dbPort != "" {
		dbConfig.Development.Port = *dbPort
	}

	// check if required database parameters exist in YAML or flags
	if dbConfig.Development.Database == "" || dbConfig.Development.User == "" {
		fmt.Println("Missing required database configuration: username or database name")
		os.Exit(1)
	}

	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbConfig.Development.Host,
		dbConfig.Development.Port,
		dbConfig.Development.Database,
		dbConfig.Development.User,
		dbConfig.Development.Password,
		*dbSSL)

	// set up application configuration and logging
	app.InProduction = *inProduction
	app.UseCache = *useCache

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// set up session management
	session = scs.New() // Using := instead of = would cause variable shadowing, as session is declared outside this function
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true                  // persist session data after the browser is closed
	session.Cookie.SameSite = http.SameSiteLaxMode // define cookie SameSite policy
	session.Cookie.Secure = app.InProduction       // insist the cookie to be encrypted - set true in production, for https

	app.Session = session

	// connect to db
	log.Println("Connecting to database... ")

	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connet to database! Dying...")
	}
	// do not defer db.SQL.Close() here; it should be in the main function
	log.Println("Connected to database")

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// create and load the template cache
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
