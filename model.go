package main

import (
	"github.com/OlegSchmidt/soup"
)

// AppReview stores all information related to a review found on the Google Play Page
type AppReview struct {
	ReviewID    string   `json:"review_id"`
	PackageName string   `json:"package_name"`
	Author      string   `json:"author"`
	Date        int64    `json:"date_posted"`
	Rating      int      `json:"rating"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	PermaLink   string   `json:"perma_link"`
	Errors      []string `json:"errors"`
}

// struct for the response
type AppReviewResponse struct {
	Status  int         `json:"status"`
	Error   string      `json:"error"`
	Reviews []AppReview `json:"reviews"`
}

func (review AppReview) fillBySoup(packageName string, documentFull soup.Root, documentReview soup.Root) AppReview {
	var lastError error = nil
	review.PackageName = packageName

	review.Author, lastError = getHtmlReviewAuthor(documentReview)
	if lastError != nil {
		review.Errors = append(review.Errors, lastError.Error())
	}

	review.Date, lastError = getHtmlReviewDate(documentReview)
	if lastError != nil {
		review.Errors = append(review.Errors, lastError.Error())
	}

	review.Rating, lastError = getHtmlReviewRating(documentReview)
	if lastError != nil {
		review.Errors = append(review.Errors, lastError.Error())
	}

	review.Title, lastError = getHtmlReviewTitle(documentReview)
	if lastError != nil {
		review.Errors = append(review.Errors, lastError.Error())
	}

	review.Body, lastError = getHtmlReviewBody(documentReview)
	if lastError != nil {
		review.Errors = append(review.Errors, lastError.Error())
	}

	return review
}
