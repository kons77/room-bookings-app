package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kons77/room-bookings-app/internal/models"
)

/*
// postData represents key-value pairs for form input testing
type postData struct {
	key   string
	value string
}
*/

// theTests contains table-driven test cases for handler testing
// Currently commented out as we're using different testing approach
var theTests = []struct {
	name               string // test name
	url                string //routes path
	method             string // get or post
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},

	/*
		{"post-search-avail", "/search-availability", "POST", []postData{
			{key: "start", value: "2024-01-01"},
			{key: "end", value: "2024-01-02"},
		}, http.StatusOK},
		{"post-search-avail-json", "/search-availability-json", "POST", []postData{
			{key: "start", value: "2024-01-01"},
			{key: "end", value: "2024-01-02"},
		}, http.StatusOK},
		{"make-reservation-post", "/make-reservation", "POST", []postData{
			{key: "first_name", value: "Joe"},
			{key: "last_name", value: "Joyson"},
			{key: "email", value: "jj@here.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusOK},
		{"make-reservation-summary", "/reservation-summary", "GET", []postData{}, http.StatusOK},
	*/
}

// TestHandlers runs table-driven tests for all handlers
// Uses httptest.NewTLSServer to simulate HTTPS connections
func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes) // ts test server
	defer ts.Close()                    // defer = close all of this after function is finished

	// Iterate through test cases
	for _, e := range theTests {
		if e.method == "GET" {
			// Handle GET request tests
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			// Verify status code matches expected
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s expect %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}

// TestRepository_Reservation tests the Reservation handler
// specifically focusing on session handling // keep models.Reservation out of the session and get it out and put it in the session
func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil) // Create new HTTP request
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder() // response recorder
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	// manually calling this - do not need routes at all for this test
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation hendler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusOK)
	}

	//test case where reservation is not in the session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation hendler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// simulate the case where trying to get a non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 1000
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation hendler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

// TestRepository_PostReservation tests the PostReservation handler
func TestRepository_PostReservation(t *testing.T) {
	layout := "2006-01-02"
	sd, _ := time.Parse(layout, "2050-01-01")
	ed, _ := time.Parse(layout, "2050-01-02")
	reservation := models.Reservation{
		RoomID:    1,
		StartDate: sd,
		EndDate:   ed,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	//	reqBody := "start_date=2050-01-01"
	//	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")  ... and go on

	/* Set и Add имеют разное поведение при работе с параметрами:
	Add:
		Добавляет новое значение к существующим значениям для данного ключа
		Позволяет иметь несколько значений для одного ключа (например, для чекбоксов или мульти-селектов)
	Set:
		Заменяет все существующие значения для данного ключа на новое значение
		Если значение для ключа уже существует, оно будет перезаписано
	*/

	postedData := url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Joe")
	postedData.Add("email", "jo@jo.com")
	postedData.Add("phone", "555-555-5555")
	postedData.Add("room_id", "1")
	encodedDate := postedData.Encode()

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(encodedDate))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	session.Put(ctx, "reservation", reservation)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusSeeOther)
	}

	// test for missing post body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted  %d", rr.Code, http.StatusSeeOther)
	}

	// Test for invalid Form
	postedData = url.Values{}
	postedData.Add("first_name", "a")
	postedData.Add("last_name", "b")
	postedData.Add("email", "c")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	session.Put(ctx, "reservation", reservation)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusSeeOther)
	}

	// Test when session is not set with reservation
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted  %d", rr.Code, http.StatusSeeOther)
	}

	// Test for failure to insert reservation into db
	postedData = url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Joe")
	postedData.Add("email", "jo@jo.com")
	postedData.Add("phone", "555-555-5555")
	postedData.Add("room_id", "2")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	reservation.RoomID = 2

	session.Put(ctx, "reservation", reservation)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed when trying to fail inserting reservation: got %d, wanted  %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test when unable to insert room restrictions
	postedData = url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Joe")
	postedData.Add("email", "jo@jo.com")
	postedData.Add("phone", "555-555-5555")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	reservation.RoomID = 1000

	session.Put(ctx, "reservation", reservation)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusTemporaryRedirect)
	}

}

func TestRepository_AvailabilityJSON(t *testing.T) {

	// room is not available
	reqBody := "start=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	// create request
	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody)) // Create new HTTP request

	// get context with session
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "x-www-form-urlencoded")

	// get response recorder
	rr := httptest.NewRecorder()

	//session.Put(ctx, "", )

	// make handler handlerfunc
	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	// make request to our handler
	handler.ServeHTTP(rr, req)

	var j jsonResponse
	err := json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	// ...........

	// ???
	if rr.Code != http.StatusOK {
		t.Errorf("Reservation hendler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusOK)
	}
}

// getCtx creates a context with session support for testing
func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
