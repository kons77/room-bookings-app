package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/kons77/room-bookings-app/internal/config"
	"golang.org/x/crypto/bcrypt"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

// Client error
func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}

func HashPassword(pswd string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pswd), 12)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}
