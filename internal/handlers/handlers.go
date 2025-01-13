package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kons77/room-bookings-app/internal/config"
	"github.com/kons77/room-bookings-app/internal/driver"
	"github.com/kons77/room-bookings-app/internal/forms"
	"github.com/kons77/room-bookings-app/internal/helpers"
	"github.com/kons77/room-bookings-app/internal/models"
	"github.com/kons77/room-bookings-app/internal/render"
	"github.com/kons77/room-bookings-app/internal/repository"
	"github.com/kons77/room-bookings-app/internal/repository/dbrepo"
)

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

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and display form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// get reservation variable from the session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("cannot get reservation from the session"))
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	res.Room.RoomName = room.RoomName

	// may change format to another one like Thursday, the 2th of January
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handlers the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// yyyy-mm-dd date format in the web form
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	startDate, err := parseDate(sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := parseDate(ed) // 	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3) // I don't agree with length 3 - Ng - chineese last name
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		// render form right from here
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// need to insert the reservation restrictions
	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1, // just for now
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability renders the availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start") // name of form input in ""
	end := r.Form.Get("end")

	startDate, err := parseDate(start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := parseDate(end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	for _, i := range rooms {
		m.App.InfoLog.Println("ROOM:", i.ID, i.RoomName)
	}

	if len(rooms) == 0 {
		// no availability
		// m.App.InfoLog.Println("No availability")
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	type RoomInfo struct {
		Image       string
		Description string
	}

	//roomInfo to hold image and description for rooms by roomID - don't want to add this info to db for now
	roomInfo := map[int]RoomInfo{
		1: {Image: "generals-quarters.png",
			Description: "A sanctuary of strength and wisdom, The General's Quarters exudes an air of commanding authority, with its rich mahogany tones and relics of triumph that seem to echo the weight of history. Here, beneath the glow of a steadfast hearth, every detail invites you to reflect upon the grandeur of leadership and the resolve of a steadfast soul.",
		},
		2: {Image: "marjors-suite.png",
			Description: "The Major's Suite is a haven of quiet introspection, where soft twilight hues and the faint scent of aged leather create an atmosphere of serene reverie. This tranquil retreat, adorned with silken drapery and the whispers of forgotten stories, offers rest to those seeking both peace and inspiration.",
		},
	}
	data["roomInfo"] = roomInfo

	// new reservation
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// store res wit start and end dates in the session to put to the next page
	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})

}

// jsonResponse
type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability and sends JSON response.
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      false,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contacts renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id")) // Atoi = ASCII to Integer
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// get that reservation variable I stored in the session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther) // code 303

}

// parseDate parses date string using the specified layout
func parseDate(dateStr string) (time.Time, error) {
	// 01/02 03:04:05PM '06 -0700  go time format
	layout := "2006-01-02" // string describes my date format yyyy-mm-dd
	return time.Parse(layout, dateStr)
}
