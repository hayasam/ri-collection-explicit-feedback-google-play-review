package main

import (
	"fmt"
	"log"
	"net/http"

	"encoding/json"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	log.SetOutput(os.Stdout)

	router := mux.NewRouter()
	router.HandleFunc("/hitec/crawl/app-reviews/google-play/{package_name}/limit/{limit}", getAppReviews).Methods("GET")

	log.Fatal(http.ListenAndServe(":9621", router))
}

func recoverAPICall(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Println("recovered from ", r)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(`{"message": "could not retrieve app reviews"}`)
	}
}

func getAppReviews(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Reviews")
	defer recoverAPICall(w)

	// get request param
	params := mux.Vars(r)
	packageName := params["package_name"]
	limit, err := strconv.Atoi(params["limit"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// crawl app reviews
	appReviews := Crawl(packageName, limit)

	// write the response
	w.Header().Set("Content-Type", "application/json")
	if len(appReviews) > 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(appReviews)
	}
}
