package scraper

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"house-pricer/olx"
	"house-pricer/otodom"
	"house-pricer/types"

	"github.com/PuerkitoBio/goquery"
)

func FetchAll() []types.Offert {
	offertsChannel := make(chan []types.Offert)
	doc := getDocOrPanic(home)
	lastPage := getLastPage(doc)
	result := []types.Offert{}
	pagesToHandle := lastPage

	for page := 1; page <= lastPage; page++ {
		go fetchPage(page, offertsChannel)
	}

	for pagesToHandle > 0 {
		offerts := <-offertsChannel
		result = append(result, offerts...)
		pagesToHandle -= 1
	}
	return result
}

const (
	root = "https://www.olx.pl"
	home = "https://www.olx.pl/nieruchomosci/mieszkania/sprzedaz/warszawa"
)

var TypeOptions = [...]string{olx.OffertType, otodom.OffertType}

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

func getDoc(url string, typ string, docChannel chan types.Doc, docErrChannel chan error) {
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

	doc := types.Doc{
		Doc:  goqueryDoc,
		Url:  url,
		Type: typ,
	}
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

func addHostUrl(host string, url string) string {
	if strings.Contains(url, "http") {
		return url
	} else {
		return host + url
	}
}

func getType(url string) (string, error) {
	for _, typ := range TypeOptions {
		if strings.Contains(url, typ) {
			return typ, nil
		}
	}
	return "", fmt.Errorf("this card does not match any of types")
}

func fetchPage(page int, ch chan []types.Offert) {
	docChannel := make(chan types.Doc)
	docErrChannel := make(chan error)
	offertChannel := make(chan types.Offert)
	offertErrChannel := make(chan error)

	url := fmt.Sprintf("%s/?page=%d", home, page)
	doc := getDocOrPanic(url)

	cardUrls := getCardUrls(doc)
	cardsToHandle := 0
	offerts := []types.Offert{}

	for _, cardUrl := range cardUrls {
		cardUrlWithHost := addHostUrl(root, cardUrl)
		typ, err := getType(cardUrlWithHost)
		if err != nil {
			continue
		}
		go getDoc(cardUrlWithHost, typ, docChannel, docErrChannel)
		cardsToHandle += 1
	}

	for cardsToHandle > 0 {
		select {
		case cardDoc := <-docChannel:
			if cardDoc.Type == olx.OffertType {
				go olx.GetOffert(cardDoc, offertChannel, offertErrChannel)
			} else {
				go otodom.GetOffert(cardDoc, offertChannel, offertErrChannel)
			}
		case cardErr := <-docErrChannel:
			log.Println("Error fetching card", cardErr)
		case offert := <-offertChannel:
			offerts = append(offerts, offert)
			cardsToHandle -= 1
		case offertErr := <-offertErrChannel:
			log.Println("Error fetching offert", offertErr)
			cardsToHandle -= 1
		}
	}
	ch <- offerts
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
