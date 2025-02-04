package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kons77/room-bookings-app/internal/driver"
	"github.com/kons77/room-bookings-app/internal/models"
)

// theTests contains table-driven test cases for handler testing
var testGET = []struct {
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
	{"non-existent", "/green/eggs/and/ham", "GET", http.StatusNotFound},
	// new routes
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/login", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new res", "/admin/reservations/new", "GET", http.StatusOK},
	{"all res", "/admin/reservations/all", "GET", http.StatusOK},
	{"cal", "/admin/reservations/cal", "GET", http.StatusOK},
	{"show res", "/admin/reservations/new/1/show", "GET", http.StatusOK},
}

// TestHandlers runs table-driven tests for all GET handlers
func TestGetHandlers(t *testing.T) {
	routes := getRoutes()
	// Uses httptest.NewTLSServer to simulate HTTPS connections
	ts := httptest.NewTLSServer(routes) // ts test server
	defer ts.Close()                    // defer = close all of this after function is finished

	// Iterate through test cases
	for _, e := range testGET {
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
	testReservation := []struct {
		name           string //test name
		resrv          models.Reservation
		resInSession   bool
		expectedStatus int
		errMessage     string
	}{
		{
			name: "everytnig is ok",
			resrv: models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			expectedStatus: http.StatusOK,
			errMessage:     "Reservation handler returned wrong response code: ",
			resInSession:   true,
		},
		{
			name:           "no reservation in the session ",
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Reservation handler returned wrong response code: ",
			resInSession:   false,
		},
		{
			name: "trying to get a non-existent room",
			resrv: models.Reservation{
				RoomID: 1000,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Reservation handler returned wrong response code: ",
			resInSession:   true,
		},
	}

	for _, e := range testReservation {
		t.Run(e.name, func(t *testing.T) {

			req, _ := http.NewRequest("POST", "/make-reservation", nil)
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			if e.resInSession {
				session.Put(ctx, "reservation", e.resrv)
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(Repo.Reservation)
			handler.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf(e.errMessage+"got %d, wanted  %d", rr.Code, e.expectedStatus)
			}
		})
	}
}

// TestRepository_PostReservation tests the PostReservation handler
func TestRepository_PostReservation(t *testing.T) {

	testPostReservation := []struct {
		name           string //test name
		postedData     url.Values
		resrv          models.Reservation
		resInSession   bool
		expectedStatus int
		errMessage     string
	}{
		{
			name: "everytnig is ok",
			postedData: url.Values{
				"start":      {"2040-01-01"},
				"end":        {"2040-01-02"},
				"first_name": {"John"},
				"last_name":  {"Joe"},
				"email":      {"jo@jo.com"},
				"phone":      {"555-555-5555"},
				"room_id":    {"1"},
			},
			resrv: models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			expectedStatus: http.StatusSeeOther,
			errMessage:     "PostReservation handler returned wrong response code when everything must be ok: ",
			resInSession:   true,
		},
		{
			name: "missing post body",
			resrv: models.Reservation{
				RoomID: 1,
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "PostReservation handler returned wrong response code: ",
			resInSession:   true,
		},
		{
			name: "invalid form",
			postedData: url.Values{
				"first_name": {"a"},
				"last_name":  {"b"},
				"email":      {"c@jo.com"},
				"room_id":    {"1"},
			},
			resrv: models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			expectedStatus: http.StatusOK,
			errMessage:     "PostReservation handler returned wrong response code: ",
			resInSession:   true,
		},
		{
			name:           "no reservation in the session",
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "PostReservation handler returned wrong response code: ",
			resInSession:   false,
		},
		{
			name: "failure to insert reservation into db",
			postedData: url.Values{
				"start":      {"2040-01-01"},
				"end":        {"2040-01-02"},
				"first_name": {"John"},
				"last_name":  {"Joe"},
				"email":      {"jo@jo.com"},
				"phone":      {"555-555-5555"},
				"room_id":    {"2"},
			},
			resrv: models.Reservation{
				RoomID: 2,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "PostReservation handler returned wrong response code: ",
			resInSession:   true,
		},
		{
			name: "failure to insert  room restrictions into db",
			postedData: url.Values{
				"start":      {"2040-01-01"},
				"end":        {"2040-01-02"},
				"first_name": {"John"},
				"last_name":  {"Joe"},
				"email":      {"jo@jo.com"},
				"phone":      {"555-555-5555"},
				"room_id":    {"1000"},
			},
			resrv: models.Reservation{
				RoomID: 1000,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "PostReservation handler returned wrong response code: ",
			resInSession:   true,
		},
	}

	for _, e := range testPostReservation {
		t.Run(e.name, func(t *testing.T) {
			var data io.Reader
			if e.postedData != nil {
				data = strings.NewReader(e.postedData.Encode())
			}

			req, _ := http.NewRequest("POST", "/make-reservation", data)
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			if e.resInSession {
				layout := "2006-01-02"
				e.resrv.StartDate, _ = time.Parse(layout, e.postedData.Get("start"))
				e.resrv.EndDate, _ = time.Parse(layout, e.postedData.Get("end"))
				session.Put(ctx, "reservation", e.resrv)
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(Repo.PostReservation)
			handler.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf(e.errMessage+"got %d, wanted  %d", rr.Code, e.expectedStatus)
			}
		})
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	// testPostAvailability contains table-driven test cases for PostAvailability handler testing
	testPostAvailability := []struct {
		name           string     //test name
		postedData     url.Values // req body
		expectedStatus int
		errMessage     string
	}{
		{
			name: "room is available",
			postedData: url.Values{
				"start": {"2040-01-01"},
				"end":   {"2040-01-02"},
			},
			expectedStatus: http.StatusOK,
			errMessage:     "Post availability when rooms ARE  available returned wrong response code",
		},
		{
			name: "room is NOT available",
			postedData: url.Values{
				"start": {"2050-01-01"},
				"end":   {"2050-01-02"},
			},
			expectedStatus: http.StatusSeeOther,
			errMessage:     "Post availability when NO rooms available returned wrong response code: ",
		},
		{
			name: "cannot query database",
			postedData: url.Values{
				"start": {"2060-01-01"},
				"end":   {"2060-01-02"},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Post availability when database query fails gave wrong status code: ",
		},
		{
			name: "invalid start date",
			postedData: url.Values{
				"start": {"invalid"},
				"end":   {"2060-01-02"},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Post availability with invalid start date gave wrong status code: ",
		},
		{
			name: "invalid end date",
			postedData: url.Values{
				"start": {"2060-01-01"},
				"end":   {"invalid"},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Post availability with invalid end date gave wrong status code: ",
		},
		{
			name:           "missing request body",
			postedData:     nil,
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Post availability with empty request body (nil) gave wrong status code:  ",
		},
	}

	// Iterate through test cases
	for _, e := range testPostAvailability {
		// must be nil if the condition is not met for missing request body test
		var data io.Reader
		if e.postedData != nil {
			data = strings.NewReader(e.postedData.Encode())
		}
		// create new request
		req, _ := http.NewRequest("POST", "/post-availability", data)

		// get context with session
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// set the request header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// get response recorder
		rr := httptest.NewRecorder()

		// make handler handlerfunc
		handler := http.HandlerFunc(Repo.PostAvailability)

		// make request to our handler
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatus {
			t.Errorf(e.errMessage+"got %d, wanted  %d", rr.Code, e.expectedStatus)
		}
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	testAvailabilityJSON := []struct {
		name        string
		postedData  url.Values
		jsonOK      bool
		jsonMessage string
		errMessage  string
	}{
		{
			name: "rooms are available",
			postedData: url.Values{
				"start":   {"2040-01-01"},
				"end":     {"2040-01-02"},
				"room_id": {"1"},
			},
			jsonOK:      true,
			jsonMessage: "",
			errMessage:  "got no availability when some was expected in AvailabilityJSON",
		},
		{
			name: "rooms are NOT available",
			postedData: url.Values{
				"start":   {"2050-01-01"},
				"end":     {"2050-01-02"},
				"room_id": {"1"},
			},
			jsonOK:      false,
			jsonMessage: "",
			errMessage:  "got availability when none was expected in AvailabilityJSON",
		},
		{
			name: "DB Error",
			postedData: url.Values{
				"start":   {"2060-01-01"},
				"end":     {"2060-01-02"},
				"room_id": {"1"},
			},
			jsonOK:      false,
			jsonMessage: "error querying database",
			errMessage:  "got true for OK when database error occurred",
		},
		{
			name:        "No request body",
			postedData:  nil,
			jsonOK:      false,
			jsonMessage: "internal server error",
			errMessage:  "got availability when request body was empty",
		},
		{
			name: "invalid start date",
			postedData: url.Values{
				"start": {"invalid"},
				"end":   {"2060-01-02"},
			},
			jsonOK:      false,
			jsonMessage: "invalid start date",
			errMessage:  "failed to handle invalid start date",
		},
		{
			name: "invalid end date",
			postedData: url.Values{
				"start": {"2060-01-01"},
				"end":   {"invalid"},
			},
			jsonOK:      false,
			jsonMessage: "invalid end date",
			errMessage:  "failed to handle invalid end date",
		},
		{
			name: "invalid room id",
			postedData: url.Values{
				"start":   {"2060-01-01"},
				"end":     {"2060-01-02"},
				"room_id": {"abc"},
			},
			jsonOK:      false,
			jsonMessage: "invalid room id",
			errMessage:  "failed to handle invalid room id",
		},
	}

	for _, tc := range testAvailabilityJSON {
		t.Run(tc.name, func(t *testing.T) {
			// must be nil if the condition is not met for missing request body test
			var data io.Reader
			if tc.postedData != nil {
				data = strings.NewReader(tc.postedData.Encode())
			}

			// create new request
			req, _ := http.NewRequest("POST", "/search-availability-json", data)

			// get context with session
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// set the request header
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// get response recorder
			rr := httptest.NewRecorder()

			// make handler handlerfunc
			handler := http.HandlerFunc(Repo.AvailabilityJSON)

			// make request to our handler
			handler.ServeHTTP(rr, req)

			var j jsonResponse
			err := json.Unmarshal(rr.Body.Bytes(), &j)
			if err != nil {
				log.Println(err)
				t.Error("failed to parse json")
			}

			if j.OK != tc.jsonOK || j.Message != tc.jsonMessage {
				t.Error(tc.errMessage)
			}
		})
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	//  set it up in the test for ChooseRoom
	// req.RequestURI = "/choose-room/1"
	testChooseRoom := []struct {
		name           string //test name
		expectedStatus int
		errMessage     string
		resrv          models.Reservation
		resInSession   bool
		urlParam       string
	}{
		{
			name:           "There's a reservation IN the session",
			expectedStatus: http.StatusSeeOther,
			errMessage:     "cannot get reservation from the session but it must be in: ",
			resrv: models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			resInSession: true,
			urlParam:     "/choose-room/1",
		},
		{
			name:           "There's NO reservation in the session",
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "get reservation from the session while it's not there: ",
			resInSession:   false,
			urlParam:       "/choose-room/1",
		},
		{
			name:           "Wrong URL parameter",
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "wrong URL parameter: ",
			resrv: models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			resInSession: true,
			urlParam:     "/choose-room/hello",
		},
	}

	for _, tc := range testChooseRoom {
		t.Run(tc.name, func(t *testing.T) {
			// create new request
			req, _ := http.NewRequest("GET", "/choose-room/1", nil)

			// get context with session
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.RequestURI = tc.urlParam

			if tc.resInSession {
				session.Put(ctx, "reservation", tc.resrv)
			}

			// set the request header
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// get response recorder
			rr := httptest.NewRecorder()

			// make handler handlerfunc
			handler := http.HandlerFunc(Repo.ChooseRoom)

			// make request to our handler
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf(tc.errMessage+"got %d, wanted  %d", rr.Code, tc.expectedStatus)
			}
		})
	}
}

func TestRepository_BookRoom(t *testing.T) {
	testBookRoom := []struct {
		name           string //test name
		expectedStatus int
		errMessage     string
		resrv          models.Reservation
		resInSession   bool
		urlParam       string
	}{
		{
			name:           "database works",
			expectedStatus: http.StatusSeeOther,
			errMessage:     "BookRoom handler returned wrong response code: ",
			resrv: models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			resInSession: true,
			urlParam:     "/book-room/?s=2040-01-01&e=2040-01-02&id=1",
		},
		{
			name:           "database fails",
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "BookRoom handler returned wrong response code: ",
			resInSession:   false,
			urlParam:       "/book-room/?s=2040-01-01&e=2040-01-02&id=4",
		},
	}

	for _, tc := range testBookRoom {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.urlParam
			// create new request
			req, _ := http.NewRequest("GET", url, nil)

			// get context with session
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.RequestURI = tc.urlParam

			if tc.resInSession {
				session.Put(ctx, "reservation", tc.resrv)
			}

			// set the request header
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// get response recorder
			rr := httptest.NewRecorder()

			// make handler handlerfunc
			handler := http.HandlerFunc(Repo.BookRoom)

			// make request to our handler
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf(tc.errMessage+"got %d, wanted  %d", rr.Code, tc.expectedStatus)
			}
		})
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	testReservationSummary := []struct {
		name           string //test name
		expectedStatus int
		errMessage     string
		resrv          models.Reservation
		resInSession   bool
	}{
		{
			name:           "There's NO reservation in the session",
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "get reservation from the session while it's not there: ",
			resInSession:   false,
		},
		{
			name:           "There's a reservation IN the session",
			expectedStatus: http.StatusOK,
			errMessage:     "cannot get reservation from the session but it must be in: ",
			resrv: models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			resInSession: true,
		},
	}

	for _, tc := range testReservationSummary {
		t.Run(tc.name, func(t *testing.T) {
			// create new request
			req, _ := http.NewRequest("POST", "/reservation-summary", nil)

			// get context with session
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			if tc.resInSession {
				session.Put(ctx, "reservation", tc.resrv)
			}

			// set the request header
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// get response recorder
			rr := httptest.NewRecorder()

			// make handler handlerfunc
			handler := http.HandlerFunc(Repo.ReservationSummary)

			// make request to our handler
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf(tc.errMessage+"got %d, wanted  %d", rr.Code, tc.expectedStatus)
			}
		})
	}
}

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string // where redtirect to
}{
	{
		"valid-credentials",
		"me@here.com", // valid email in the test-repo.go
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"jack@nimble.com",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invaid-data",
		"j",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

func Test_Login(t *testing.T) {
	// range through all tests
	for _, tc := range loginTests {
		t.Run(tc.name, func(t *testing.T) {
			postedData := url.Values{
				"email":    {tc.email},
				"password": {"password"},
			}
			// create new request
			req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
			// get context with session
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			// set the request header
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			// get response recorder
			rr := httptest.NewRecorder()
			// make handler handlerfunc
			handler := http.HandlerFunc(Repo.PostShowLogin)
			// make request to our handler
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("failed %s:  expected code %d, but got %d", tc.name, tc.expectedStatusCode, rr.Code)
			}

			if tc.expectedLocation != "" {
				// get the url from test
				actualLoc, _ := rr.Result().Location()
				if actualLoc.String() != tc.expectedLocation {
					t.Errorf("failed %s:  expected location %s, but got location %s",
						tc.name, tc.expectedLocation, actualLoc.String())
				}
			}

			// checking for expected values in HTML - never fire
			if tc.expectedHTML != "" {
				// read the response body into a string
				html := rr.Body.String()
				if !strings.Contains(html, tc.expectedHTML) {
					t.Errorf("failed %s:  expected to find %s, but did not", tc.name, tc.expectedHTML)
				}
			}
		})

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
