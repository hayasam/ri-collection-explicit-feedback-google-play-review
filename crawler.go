package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/jehiah/go-strftime"
)

// common html attributes
const (
	div   = "div"
	span  = "span"
	class = "class"
	a     = "a"

	reviewsURL = "https://play.google.com/store/getreviews?authuser=0"
)

// Crawl crawls the reviews of a given app until a given limit is reached
func Crawl(packageName string, limit int) []AppReview {
	var appReviews []AppReview

	page := 0

	for {
		page++
		// sleep for 6 seconds to not be blocked by Google
		//time.Sleep(6 * time.Second)

		// request html page
		formData := url.Values{}
		formData.Add("reviewType", "0")
		formData.Add("pageNum", strconv.Itoa(page))
		formData.Add("id", packageName)
		formData.Add("reviewSortOrder", "0")
		formData.Add("xhr", "1")
		formData.Add("hl", "en")

		var resp *http.Response
		var err error
		resp, err = http.PostForm(reviewsURL, formData)
		if err != nil || resp == nil {
			fmt.Printf("%s ERROR: %s\n", packageName, err)
			return appReviews
		}
		// handle exit strategies
		code := resp.StatusCode
		if code == 400 || code == 403 || code == 404 || code == 408 || code == 429 {
			fmt.Printf("%s STATUS %d: no more reviews\n", packageName, code)
			return appReviews
		}

		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s ERROR: %s\n", packageName, err)
			return appReviews
		}
		resp.Body.Close()

		// pre-process the html
		stringContent := escapedBytesToString(contents)

		// extract data from reviews of the html
		doc := soup.HTMLParse(stringContent)

		// check if the captcha came up
		captcha := doc.Find("body").Attrs()["onload"]
		if captcha == "e=document.getElementById('captcha');if(e){e.focus();}" {
			fmt.Printf("%s QUIT PROGRAMM: captcha needed\n", packageName)
			return appReviews
		}

		var reviewsOnPage int
		reviewsInPage := doc.FindAll(div, class, "single-review")
		for _, rDoc := range reviewsInPage {
			review := AppReview{}
			review.PackageName = packageName
			// review.Title = getReviewTitle(rDoc)
			review.Body = getReviewBody(rDoc)
			review.Date = getReviewDate(rDoc)
			review.Author = getReviewAuthor(rDoc)
			review.PermaLink = getReviewPermaLink(rDoc)
			review.ReviewID = getReviewID(rDoc)
			review.Rating = getReviewRating(rDoc)

			reviewsOnPage++
			appReviews = append(appReviews, review)

			if limit > 0 && len(appReviews) == limit {
				break
			}
		}

		if reviewsOnPage == 0 { // no more reviews
			break
		}

		break
	}
	return appReviews
}

func getReviewTitle(doc soup.Root) string {
	return doc.Find(span, class, "review-title").Text()
}

func getReviewBody(doc soup.Root) string {
	return doc.Find(span, class, "review-title").FindNextSibling().NodeValue
}

func getReviewDate(doc soup.Root) int64 {
	unFormattedDate := doc.Find(span, class, "review-date").Text()
	t, err := time.Parse("January 2, 2006", unFormattedDate)

	if err != nil {
		return -1
	}

	s := strftime.Format("%Y%m%d", t)
	val, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}

	return int64(val)
}

func getReviewAuthor(doc soup.Root) string {
	return strings.TrimSpace(doc.Find(span, class, "author-name").Text())
}

func getReviewPermaLink(doc soup.Root) string {
	return "https://play.google.com" + doc.Find(a, class, "reviews-permalink").Attrs()["href"]
}

func getReviewID(doc soup.Root) string {
	return doc.Find(div, class, "review-header").Attrs()["data-reviewid"]
}

func getReviewRating(doc soup.Root) int {
	ratingRaw := doc.Find(div, class, "current-rating").Attrs()["style"]
	re := regexp.MustCompile("[^0-9]+")
	i, err := strconv.Atoi(re.ReplaceAllString(ratingRaw, ""))
	if err != nil {
		fmt.Println(err)
	} else {
		if i == 20 {
			return 1
		} else if i == 40 {
			return 2
		} else if i == 60 {
			return 3
		} else if i == 80 {
			return 4
		} else if i == 100 {
			return 5
		}
	}
	return -1
}

func getHelpfulness(doc soup.Root) int {
	fmt.Println("get the helpfulnes score")
	helpfulnessScoreRaw := doc.Find(div, "aria-label", "Number of times this review was rated helpful").Text()
	re := regexp.MustCompile("[^0-9]+")
	i, err := strconv.Atoi(re.ReplaceAllString(helpfulnessScoreRaw, ""))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Print("\nHelpfulness")
	fmt.Println(i)

	return i
}

func escapedBytesToString(b []byte) string {
	b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
	b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	b = bytes.Replace(b, []byte("\\u003d"), []byte("="), -1)
	b = bytes.Replace(b, []byte("\\\""), []byte("\""), -1)
	return string(b)
}
