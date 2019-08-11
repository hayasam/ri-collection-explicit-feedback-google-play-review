package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/OlegSchmidt/soup"
	"github.com/jehiah/go-strftime"
)

// common html attributes
const (
	// common html tags
	div    = "div"
	button = "button"
	span   = "span"
	class  = "class"
	a      = "a"
	meta   = "meta"

	// common html attributes
	attributeClass    = "class"
	attributeProperty = "property"
	attributeContent  = "content"
	attributeRole     = "role"
	attributeJsname   = "jsname"

	// string constants
	propertyValueOpengraphUrl = "og:url"
	attributeValueImg         = "img"

	// css classes for selection
	classContentApp           = "LXrl4c"
	classMainContentBlock     = "W4P4ne"
	classReviewAreasContainer = "d15Mdf"
	classReviewAuthor         = "X43Kjb"
	classReviewDate           = "p2TkOb"
	classReviewStarFilled     = "vQHuPe"
	classReviewTitle          = "IEFhEe"

	// jsname values for selection
	jsnameReviewContentShort = "bN97Pc"
	jsnameReviewContentFull  = "fbQN7e"

	reviewsURL = "https://play.google.com/store/getreviews?authuser=0"
)

// parses the website and returns the DOM struct
func retrieveDoc(url string) (soup.Root, int) {
	var document soup.Root
	httpStatus := http.StatusOK
	// retrieving the html page
	response, soupError := soup.Get(url)
	if soupError != nil {
		fmt.Println("\tcould not reach", url, "because of the following error:")
		fmt.Println(soupError)
		httpStatus = http.StatusBadRequest
	} else {
		// pre-process html
		response = strings.Replace(response, "<br>", "\n", -1)
		response = strings.Replace(response, "<b>", "", -1)
		response = strings.Replace(response, "</b>", "", -1)
		document = soup.HTMLParse(response)
	}

	return document, httpStatus
}

// returns the container where the reviews are stored
func GetReviewContainer(document soup.Root) (soup.Root, error) {
	var container soup.Root
	var containerError error = nil
	appContainers := document.FindAll(div, class, classContentApp)
	if len(appContainers) >= 1 {
		mainContentBlock := appContainers[len(appContainers)-1].Find(div, class, classMainContentBlock)
		if mainContentBlock.Error == nil {
			mainContentBlockChildren := mainContentBlock.Children()
			if len(mainContentBlockChildren) >= 2 {
				containerBlockReviewChildren := mainContentBlockChildren[1].Children()
				if len(containerBlockReviewChildren) >= 3 {
					container = containerBlockReviewChildren[2]
				} else {
					containerError = errors.New("2nd child of main container block for reviews should contain at least 3 children")
				}
			} else {
				containerError = errors.New("main container block for reviews (2nd main content block) should contain at least 2 children")
			}
		} else {
			containerError = errors.New("couldn't find the main content blocks in the main container, looking for first <div class=\"" + classMainContentBlock + "\"></div>")
		}
	} else {
		containerError = errors.New("couldn't find the main container of the app, looking for last <div class=\"" + classContentApp + "\"></div>")
	}

	return container, containerError
}

// returns the 3 main areas of the review : headline (stars, name, date), review itself and the reply from developer
func getReviewAreas(document soup.Root) ([]soup.Root, error) {
	var reviewAreas []soup.Root
	var reviewAreasError error = nil

	areaContainer := document.Find(div, class, classReviewAreasContainer)
	if areaContainer.Error == nil {
		areaContainerChildren := areaContainer.Children()
		if len(areaContainerChildren) >= 2 {
			reviewAreas = areaContainerChildren
		} else {
			reviewAreasError = errors.New("mode \"html\" : <div class=\"" + classReviewAreasContainer + "\"></div> should contain at least 2 children")
		}
	} else {
		reviewAreasError = errors.New("mode \"html\" : couldn't find container for the 3 areas of the review, looking for <div class=\"" + classReviewAreasContainer + "\"></div>")
	}

	return reviewAreas, reviewAreasError
}

// crawls the given link assuming that there are reviews to be found
func CrawlHtml(link string, arguments ...bool) ([]AppReview, error) {
	var appReviews []AppReview
	var crawlError error = nil
	var review AppReview

	appPage, HttpStatus := retrieveDoc(link)
	if HttpStatus == http.StatusOK {
		packageName, packageNameError := getHtmlReviewPackageName(appPage)
		if packageNameError == nil {
			reviewBlock, reviewBlockError := GetReviewContainer(appPage)
			if reviewBlockError == nil {
				reviewElements := reviewBlock.Children()
				for position := range reviewElements {
					review = AppReview{}.fillBySoup(packageName, appPage, reviewElements[position])
					appReviews = append(appReviews, review)
				}
			} else {
				crawlError = reviewBlockError
			}
		}
	} else {
		crawlError = errors.New("given link couldn't be parsed, please check the online availability")
	}

	return appReviews, crawlError
}

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
		captcha := doc.Find("body").GetAttribute("onload")
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
	return "https://play.google.com" + doc.Find(a, class, "reviews-permalink").GetAttribute("href")
}

func getReviewID(doc soup.Root) string {
	return doc.Find(div, class, "review-header").GetAttribute("data-reviewid")
}

func getReviewRating(doc soup.Root) int {
	ratingRaw := doc.Find(div, class, "current-rating").GetAttribute("style")
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
