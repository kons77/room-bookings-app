package handlers

import (
	"context"
	"encoding/json"
	"io"
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
// postedData represents key-value pairs for form input testing
type postedData struct {
	key   string
	value string
}
*/

// theTests contains table-driven test cases for handler testing
// Currently commented out as we're using different testing approach
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

	/*
		{"post-search-avail", "/search-availability", "POST", []postedData{
			{key: "start", value: "2024-01-01"},
			{key: "end", value: "2024-01-02"},
		}, http.StatusOK},
		{"post-search-avail-json", "/search-availability-json", "POST", []postedData{
			{key: "start", value: "2024-01-01"},
			{key: "end", value: "2024-01-02"},
		}, http.StatusOK},
		{"make-reservation-post", "/make-reservation", "POST", []postedData{
			{key: "first_name", value: "Joe"},
			{key: "last_name", value: "Joyson"},
			{key: "email", value: "jj@here.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusOK},
		{"make-reservation-summary", "/reservation-summary", "GET", []postedData{}, http.StatusOK},
	*/
}

// TestHandlers runs table-driven tests for all handlers
// Uses httptest.NewTLSServer to simulate HTTPS connections
func TestHandlers(t *testing.T) {
	routes := getRoutes()
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
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusOK)
	}

	//test case where reservation is not in the session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusTemporaryRedirect)
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
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted  %d", rr.Code, http.StatusTemporaryRedirect)
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

	/*  also works this way:
	postedData := url.Values{
		"first_name": []string{"John"},
		"last_name": []string{"Joe"},
	}
	*/
	postedData := url.Values{
		"first_name": []string{"John"},
		"last_name":  []string{"Joe"},
		"email":      []string{"jo@jo.com"},
		"phone":      []string{"555-555-5555"},
		"room_id":    []string{"1"},
	}
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
	postedData = url.Values{
		"first_name": []string{"a"},
		"last_name":  []string{"b"},
		"email":      []string{"c"},
		"room_id":    []string{"1"},
	}

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
	postedData = url.Values{
		"first_name": []string{"John"},
		"last_name":  []string{"Joe"},
		"email":      []string{"jo@jo.com"},
		"phone":      []string{"555-555-5555"},
		"room_id":    []string{"2"},
	}

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
	postedData = url.Values{
		"first_name": []string{"John"},
		"last_name":  []string{"Joe"},
		"email":      []string{"jo@jo.com"},
		"phone":      []string{"555-555-5555"},
		"room_id":    []string{"1"},
	}

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
				"start": []string{"2040-01-01"},
				"end":   []string{"2040-01-02"},
			},
			expectedStatus: http.StatusOK,
			errMessage:     "Post availability when rooms ARE  available returned wrong response code",
		},
		{
			name: "room is NOT available",
			postedData: url.Values{
				"start": []string{"2050-01-01"},
				"end":   []string{"2050-01-02"},
			},
			expectedStatus: http.StatusSeeOther,
			errMessage:     "Post availability when NO rooms available returned wrong response code: ",
		},
		{
			name: "cannot query database",
			postedData: url.Values{
				"start": []string{"2060-01-01"},
				"end":   []string{"2060-01-02"},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Post availability when database query fails gave wrong status code: ",
		},
		{
			name: "invalid start date",
			postedData: url.Values{
				"start": []string{"invalid"},
				"end":   []string{"2060-01-02"},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			errMessage:     "Post availability with invalid start date gave wrong status code: ",
		},
		{
			name: "invalid end date",
			postedData: url.Values{
				"start": []string{"2060-01-01"},
				"end":   []string{"invalid"},
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
				"start":   []string{"2040-01-01"},
				"end":     []string{"2040-01-02"},
				"room_id": []string{"1"},
			},
			jsonOK:      true,
			jsonMessage: "",
			errMessage:  "got no availability when some was expected in AvailabilityJSON",
		},
		{
			name: "rooms are NOT available",
			postedData: url.Values{
				"start":   []string{"2050-01-01"},
				"end":     []string{"2050-01-02"},
				"room_id": []string{"1"},
			},
			jsonOK:      false,
			jsonMessage: "",
			errMessage:  "got availability when none was expected in AvailabilityJSON",
		},
		{
			name: "DB Error",
			postedData: url.Values{
				"start":   []string{"2060-01-01"},
				"end":     []string{"2060-01-02"},
				"room_id": []string{"1"},
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
				"start": []string{"invalid"},
				"end":   []string{"2060-01-02"},
			},
			jsonOK:      false,
			jsonMessage: "invalid start date",
			errMessage:  "failed to handle invalid start date",
		},
		{
			name: "invalid end date",
			postedData: url.Values{
				"start": []string{"2060-01-01"},
				"end":   []string{"invalid"},
			},
			jsonOK:      false,
			jsonMessage: "invalid end date",
			errMessage:  "failed to handle invalid end date",
		},
		{
			name: "invalid room id",
			postedData: url.Values{
				"start":   []string{"2060-01-01"},
				"end":     []string{"2060-01-02"},
				"room_id": []string{"abc"},
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
			req, _ := http.NewRequest("POST", "/post-availability", data)

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

}

// getCtx creates a context with session support for testing
func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
