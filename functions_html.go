package main

import (
	"errors"
	"github.com/OlegSchmidt/soup"
	"github.com/jehiah/go-strftime"
	"strconv"
	"strings"
	"time"
)

// returns the name of the package
func getHtmlReviewPackageName(document soup.Root) (string, error) {
	packageName := ""
	var packageNameError error = nil

	appUrlMeta := document.Find(meta, attributeProperty, propertyValueOpengraphUrl)
	if appUrlMeta.Error == nil && appUrlMeta.HasAttribute(attributeContent) {
		appLink := appUrlMeta.GetAttribute(attributeContent)
		appLinkParts := strings.Split(appLink, "?")
		if len(appLinkParts) == 2 {
			getParameter := appLinkParts[1]
			parameters := strings.Split(getParameter, "&")
			for parameterPosition := range parameters {
				parameterParts := strings.Split(parameters[parameterPosition], "=")
				if len(parameterParts) == 2 {
					if parameterParts[0] == "id" {
						packageName = parameterParts[1]
					}
				}
			}
			if packageName == "" {
				packageNameError = errors.New("mode \"html\" property \"packageName\" : query parameter \"id\" of the app url is empty")
			}
		} else {
			packageNameError = errors.New("mode \"html\" property \"packageName\" : app url should contain query-parameter")
		}
	} else {
		packageNameError = errors.New("mode \"html\" property \"packageName\" : cannot find meta of the app to parse package name, looking for <meta property=\"" + propertyValueOpengraphUrl + "\" content=\"\"></meta>")
	}

	return packageName, packageNameError
}

// returns the review author
func getHtmlReviewAuthor(document soup.Root) (string, error) {
	author := ""
	var authorError error = nil

	contentAreas, contentAreasError := getReviewAreas(document)
	if contentAreasError == nil {
		authorBlock := contentAreas[0].Find(span, attributeClass, classReviewAuthor)
		if authorBlock.Error == nil {
			authorName := authorBlock.Text()
			if authorName != "" {
				author = authorName
			} else {
				authorError = errors.New("mode \"html\" property \"author\" : 1st child of headline area (left part) is empty")
			}
		} else {
			authorError = errors.New("mode \"html\" property \"author\" : author element not found, looking for <span class=\"" + classReviewAuthor + "\"></span>")
		}
	} else {
		authorError = contentAreasError
	}

	return author, authorError
}

// returns the date when the review was made
func getHtmlReviewDate(document soup.Root) (int64, error) {
	var date int64 = -1
	var dateError error = nil
	monthMap := map[string]string{
		"Januar":    "January",
		"Februar":   "February",
		"März":      "March",
		"April":     "April",
		"Mai":       "May",
		"Juni":      "June",
		"Juli":      "July",
		"August":    "August",
		"September": "September",
		"Oktober":   "October",
		"November":  "November",
		"Dezember":  "December",
	}

	contentAreas, contentAreasError := getReviewAreas(document)
	if contentAreasError == nil {
		dateBlock := contentAreas[0].Find(span, attributeClass, classReviewDate)
		if dateBlock.Error == nil {
			dateString := dateBlock.Text()
			if dateString != "" {
				for monthGerman, monthEnglish := range monthMap {
					dateString = strings.Replace(dateString, monthGerman, monthEnglish, -1)
				}
				dateParsed, dateParsedError := time.Parse("2. January 2006", dateString)
				if dateParsedError == nil {
					dateFormatted := strftime.Format("%Y%m%d", dateParsed)
					dateNumber, dateNumberError := strconv.ParseInt(dateFormatted, 0, 64)
					if dateNumberError == nil {
						date = dateNumber
					} else {
						dateError = errors.New("mode \"html\" property \"date\" : date couldn't be formatted to int64")
					}
				} else {
					dateError = errors.New("mode \"html\" property \"date\" : date couldn't be parsed")
				}
			} else {
				dateError = errors.New("mode \"html\" property \"date\" : 2nd span of 2nd child of headline area (left part) is empty")
			}
		} else {
			dateError = errors.New("mode \"html\" property \"date\" : date element not found, looking for <span class=\"" + classReviewDate + "\"></span>")
		}
	} else {
		dateError = contentAreasError
	}

	return date, dateError
}
func getHtmlReviewRating(document soup.Root) (int, error) {
	var rating int = 0
	var ratingError error = nil

	contentAreas, contentAreasError := getReviewAreas(document)
	if contentAreasError == nil {
		ratingBlock := contentAreas[0].Find(div, attributeRole, attributeValueImg)
		if ratingBlock.Error == nil {
			ratingStarsMarked := len(ratingBlock.FindAll(div, attributeClass, classReviewStarFilled))
			if ratingStarsMarked > 0 {
				rating = ratingStarsMarked
			} else {
				ratingError = errors.New("mode \"html\" property \"rating\" : <div class=\"" + classReviewStarFilled + "\"></div> is marking a filled star but there are none, please check the CSS-class")
			}
		} else {
			ratingError = errors.New("mode \"html\" property \"rating\" : date element not found, looking for <div role=\"" + attributeValueImg + "\"></div>")
		}
	} else {
		ratingError = contentAreasError
	}

	return rating, ratingError
}

// returns the title of the review, if its available
func getHtmlReviewTitle(document soup.Root) (string, error) {
	title := ""
	var titleError error = nil

	contentAreas, contentAreasError := getReviewAreas(document)
	if contentAreasError == nil {
		titleBlock := contentAreas[1].Find(span, attributeClass, classReviewTitle)
		if titleBlock.Error == nil && titleBlock.Text() != "" {
			title = titleBlock.Text()
		}
	} else {
		titleError = contentAreasError
	}

	return title, titleError
}

// returns the title of the review, if its available
func getHtmlReviewBody(document soup.Root) (string, error) {
	body := ""
	var bodyError error = nil

	contentAreas, contentAreasError := getReviewAreas(document)
	if contentAreasError == nil {
		reviewContentShort := contentAreas[1].Find(span, attributeJsname, jsnameReviewContentShort)
		reviewContentFull := contentAreas[1].Find(span, attributeJsname, jsnameReviewContentFull)
		if reviewContentShort.Error == nil {
			if reviewContentShort.Find(button).Error == nil {
				if reviewContentFull.Error == nil {
					reviewText := reviewContentFull.Text()
					if reviewText != "" {
						body = reviewText
					}
				} else {
					bodyError = errors.New("mode \"html\" property \"body\" : the review was shortened but cant find the full review text")
				}
			}
			if body == "" {
				reviewText := reviewContentShort.Text()
				if reviewText != "" {
					body = reviewText
				} else {
					bodyError = errors.New("mode \"html\" property \"body\" : cannot find the review text")
				}
			}
		} else {
			bodyError = errors.New("mode \"html\" property \"body\" : short review text block not found, looking for <span jsname=\"" + jsnameReviewContentShort + "\"></span>")
		}
	} else {
		bodyError = contentAreasError
	}

	return body, bodyError
}
