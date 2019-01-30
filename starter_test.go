package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

var router *mux.Router

func TestMain(m *testing.M) {
	fmt.Println("--- Start Tests")
	setup()

	// run the test cases defined in this file
	retCode := m.Run()

	tearDown()

	// call with result of m.Run()
	os.Exit(retCode)
}

func setup() {
	fmt.Println("--- --- setup")
	setupRouter()
}

func setupRouter() {
	router = mux.NewRouter()
	// Insert
	router.HandleFunc("/hitec/crawl/app-reviews/google-play/{package_name}/limit/{limit}", getAppReviews).Methods("GET")
}

func tearDown() {
	fmt.Println("--- --- tear down")
}

func buildRequest(method, endpoint string, payload io.Reader, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, endpoint, payload)
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	return req
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}
func TestGetAppReviews(t *testing.T) {
	fmt.Println("start TestGetAppReviewsOfClass")
	var method = "GET"
	var endpoint = "/hitec/crawl/app-reviews/google-play/%s/limit/%s"

	/*
	 * test for success CHECK 1
	 */
	endpointCheckOne := fmt.Sprintf(endpoint, "com.whatsapp", "30")
	req := buildRequest(method, endpointCheckOne, nil, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var appReviews []AppReview
	err := json.NewDecoder(rr.Body).Decode(&appReviews)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(appReviews) != 30 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 30, len(appReviews))
	}
}
