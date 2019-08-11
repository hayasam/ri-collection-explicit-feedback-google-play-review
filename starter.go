package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"encoding/json"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	modeApi             = "api"
	modeHtml            = "html"
	errorRecover        = "Could not retrieve app reviews"
	errorParameterLimit = "Given parameter \"limit\" is not valid, it should be an integer"
)

func main() {
	parameters := os.Args
	if len(parameters) >= 3 && parameters[len(os.Args)-3] == "local" {
		mode := parameters[len(os.Args)-2]
		switch mode {
		case modeHtml:
			url := parameters[len(os.Args)-1]
			CrawlHtml(url, true)
			break
		case modeApi:
			break
		}
	} else {
		log.SetOutput(os.Stdout)

		router := mux.NewRouter()
		router.HandleFunc("/hitec/crawl/app-reviews/google-play/{package_name}/limit/{limit}", getAppReviews).Methods("GET")
		router.HandleFunc("/hitec/crawl/app-reviews/google-play/static/", getAppReviewsStatic).Methods("GET")

		log.Fatal(http.ListenAndServe(":9621", router))
	}
}

// error handling
func recoverAPICall(w http.ResponseWriter, appReviewResponse AppReviewResponse) {
	if r := recover(); r != nil {
		log.Println("recovered from ", r)
		appReviewResponse.Status = http.StatusInternalServerError
		appReviewResponse.Error = errorRecover
	}
}

// parsing of google play page to get reviews
func getAppReviews(writer http.ResponseWriter, request *http.Request) {
	appReviewResponse := AppReviewResponse{}
	appReviewResponse.Status = http.StatusOK

	fmt.Println("Get Reviews")
	defer recoverAPICall(writer, appReviewResponse)

	// get request param
	params := mux.Vars(request)
	packageName := params["package_name"]
	limit, limitError := strconv.Atoi(params["limit"])
	if limitError == nil {
		appReviewResponse.Reviews = Crawl(packageName, limit)
	} else {
		appReviewResponse.Status = http.StatusBadRequest
		appReviewResponse.Error = errorParameterLimit
	}

	serveResponse(writer, appReviewResponse)
}

// parsing of given URL to get reviews
func getAppReviewsStatic(writer http.ResponseWriter, request *http.Request) {
	appReviewResponse := AppReviewResponse{}
	appReviewResponse.Status = http.StatusOK

	fmt.Println("Get Reviews from given website")
	defer recoverAPICall(writer, appReviewResponse)

	queryParameter := request.URL.Query()

	for parameter, value := range queryParameter {
		if parameter == "target_url" {
			var crawlError error = nil
			url := strings.Join(value, "")
			appReviewResponse.Reviews, crawlError = CrawlHtml(url, true)
			if crawlError != nil {
				appReviewResponse.Status = http.StatusBadRequest
				appReviewResponse.Error = crawlError.Error()
			}
			break
		}
	}

	serveResponse(writer, appReviewResponse)
}

// serves the generated content
func serveResponse(writer http.ResponseWriter, appReviewResponse AppReviewResponse) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(appReviewResponse.Status)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(appReviewResponse)
}
