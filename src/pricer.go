package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Offert struct {
	URL         string    `json:"url" dynamodbav:"URL"`          // Partition Key, string type
	Title       string    `json:"title" dynamodbav:"Title"`
	Price       int       `json:"price" dynamodbav:"Price"`
	Area        string    `json:"area" dynamodbav:"Area"`
	Mortgage    float64   `json:"mortgage" dynamodbav:"Mortgage"`
	Furnished   int       `json:"furnished" dynamodbav:"Furnished"` // 1 true, 0 unknown, -1 false
	// Added timestamps for better record keeping.
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"CreatedAt"`   // Stored as ISO 8601 string
	LastUpdated time.Time `json:"lastUpdated" dynamodbav:"LastUpdated"` // Stored as ISO 8601 string
}

type Doc struct {
	doc *goquery.Document
	url string
}

const (
	root = "https://www.olx.pl"
	home = "https://www.olx.pl/nieruchomosci/mieszkania/sprzedaz/warszawa"

	TIMEOUT = 5
)

var (
	nonDigit = regexp.MustCompile(`\D`)
	price    = regexp.MustCompile(`[\d\.,]+`)
)

func getDocOrPanic(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic(err)
	}

	return doc
}

func getDoc(url string, docChannel chan Doc, docErrChannel chan error) {
	res, err := http.Get(url)
	if err != nil {
		docErrChannel <- err
	}
	defer res.Body.Close()

	goqueryDoc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		docErrChannel <- err
		return
	}

	doc := Doc{goqueryDoc, url}
	docChannel <- doc
}

func getCardUrls(doc *goquery.Document) []string {
	urls := []string{}
	iter := doc.Find(`div[data-cy="l-card"] a[href]`).EachIter()
	for _, element := range iter {
		url, exists := element.Attr("href")
		if exists {
			urls = append(urls, url)
		}
	}
	return urls
}

func parsePrice(price string) (int, error) {
	digits := nonDigit.ReplaceAllString(price, "")
	result, err := strconv.Atoi(digits)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func getFurnished(tag string) int {
	re := regexp.MustCompile(`Umeblowane: (\w+)</p>`)
	matches := re.FindStringSubmatch(tag)

	if len(matches) < 2 {
		return 0
	}

	switch strings.ToLower(matches[1]) {
	case "tak":
		return 1
	case "nie":
		return -1
	default:
		return 0
	}
}

func getArea(tag string) (string, error) {
	re := regexp.MustCompile(`Powierzchnia: ([\d,\.]*)\sm`)
	matches := re.FindStringSubmatch(tag)

	if len(matches) < 2 {
		return "?", errors.New("No area tag")
	}
	areaStr := strings.ReplaceAll(matches[1], ",", ".")
	result, err := strconv.ParseFloat(areaStr, 64)
	if err != nil {
		return "?", errors.New("Cant convert area to float")
	}
	return strconv.FormatFloat(result, 'f', 0, 64), nil
}

func getOffert(doc Doc, offertChannel chan Offert, offertErrChannel chan error) {
	element := doc.doc.Find(`div[data-testid="ad-price-container"] h3`).First()
	priceStr := element.Text()
	price, err := parsePrice(priceStr)
	if err != nil {
		offertErrChannel <- errors.New(doc.url + " " + err.Error())
		return
	}

	tagContainer := doc.doc.Find(`div[data-testid="ad-parameters-container"]`).First()
	text := tagContainer.Text()
	furnished := getFurnished(text)
	area, err := getArea(text)
	if err != nil {
		offertErrChannel <- errors.New(doc.url + " " + err.Error())
	}

	offert := Offert{
		URL:         doc.url,
		Title:       "",
		Price:       price,
		Area:        area,
		Mortgage:     0,
		Furnished:   furnished,
	}
	offertChannel <- offert
}

func addHostUrl(host string, url string) string {
	if strings.Contains(url, "http") {
		return url
	} else {
		return host + url
	}
}

func fetchPage(page int, ch chan []Offert) {
	docChannel := make(chan Doc)
	docErrChannel := make(chan error)
	offertChannel := make(chan Offert)
	offertErrChannel := make(chan error)

	url := fmt.Sprintf("%s/?page=%d", home, page)
	doc := getDocOrPanic(url)

	cardUrls := getCardUrls(doc)
	offerts := []Offert{}

	for _, cardUrl := range cardUrls {
		carUrlWithHost := addHostUrl(root, cardUrl)
		if strings.Contains(carUrlWithHost, "olx") {
			go getDoc(carUrlWithHost, docChannel, docErrChannel)
		}
	}

	t := time.Now().Add(TIMEOUT * time.Second)
	resetTimout := func() { t = time.Now().Add(TIMEOUT * time.Second) }
	for {
		select {
		case cardDoc := <-docChannel:
			go getOffert(cardDoc, offertChannel, offertErrChannel)
			resetTimout()
		case cardErr := <-docErrChannel:
			log.Println("Error fetching card", cardErr)
			resetTimout()
		case offert := <-offertChannel:
			offerts = append(offerts, offert)
			resetTimout()
		case offertErr := <-offertErrChannel:
			log.Println("Error fetching offert", offertErr)
			resetTimout()
		default:
			if time.Now().Before(t) {
				time.Sleep(time.Second)
				continue
			}

			ch <- offerts
			return
		}
	}
}

func getLastPage(doc *goquery.Document) int {
	wrapper := doc.Find(`[data-testid="pagination-wrapper"]`).First()
	last := wrapper.Find(`[data-testid="pagination-list-item"]`).Last()
	lastLink, exists := last.Find(`a`).First().Attr("href")

	if !exists {
		return 30
	}
	pageStr := strings.Split(lastLink, "?page=")[1]
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return 30
	}
	return page
}

func fetchAll() []Offert {
	offertsChannel := make(chan []Offert)
	doc := getDocOrPanic(home)
	lastPage := getLastPage(doc)
	result := []Offert{}

	for page := 1; page <= lastPage; page++ {
		go fetchPage(page, offertsChannel)
	}

	t := time.Now().Add(TIMEOUT * time.Second)
	resetTimout := func() { t = time.Now().Add(TIMEOUT * time.Second) }
	for {
		select {
		case offerts := <-offertsChannel:
			result = append(result, offerts...)
			resetTimout()
		default:
			if time.Now().Before(t) {
				time.Sleep(time.Second)
				continue
			}

			log.Println("Timeout exceded. Quitting...")

			return result
		}
	}
}
