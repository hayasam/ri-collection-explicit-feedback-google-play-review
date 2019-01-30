package main

// AppReview stores all information related to a review found on the Google Play Page
type AppReview struct {
	ReviewID    string `json:"review_id" bson:"review_id"`
	PackageName string `json:"package_name" bson:"package_name"`
	Author      string `json:"author" bson:"author"`
	Date        int64  `json:"date_posted" bson:"date_posted"`
	Rating      int    `json:"rating" bson:"rating"`
	Title       string `json:"title" bson:"title"`
	Body        string `json:"body" bson:"body"`
	PermaLink   string `json:"perma_link" bson:"perma_link"`
}
