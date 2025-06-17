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
	result := []types.Offert{}
	offertsChannel := make(chan []types.Offert)
	threadsWaiting := 0

	for urlWithParams := range olx.UrlWithParamsGenerator() {
		fmt.Println(urlWithParams)
		threadsWaiting += 1
		go fetchAllFromUrl(urlWithParams, offertsChannel)
	}

	for threadsWaiting > 0 {
		offerts := <-offertsChannel
		result = append(result, offerts...)
		threadsWaiting -= 1
	}

	return result
}

func fetchAllFromUrl(urlWithParams string, resultChannel chan []types.Offert) {
	offertsChannel := make(chan []types.Offert)
	doc := getDocOrPanic(urlWithParams)
	lastPage := getLastPage(doc)
	result := []types.Offert{}
	pagesToHandle := lastPage

	for page := 1; page <= lastPage; page++ {
		go fetchPage(olx.Root, urlWithParams, page, offertsChannel)
	}

	for pagesToHandle > 0 {
		offerts := <-offertsChannel
		result = append(result, offerts...)
		pagesToHandle -= 1
	}
	resultChannel <- result
}

var TypeOptions = [...]string{
	olx.OffertType,
	// otodom.OffertType,
}

var (
	nonDigit = regexp.MustCompile(`\D`)
	price    = regexp.MustCompile(`[\d\.,]+`)
)

func getDocOrPanic(uri string) *goquery.Document {
	res, err := http.Get(uri)
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

func getDoc(uri string, typ string, docChannel chan types.Doc, docErrChannel chan error) {
	res, err := http.Get(uri)
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
		Url:  uri,
		Type: typ,
	}
	docChannel <- doc
}

func getCarduris(doc *goquery.Document) []string {
	uris := []string{}
	iter := doc.Find(`div[data-cy="l-card"] a[href]`).EachIter()
	for _, element := range iter {
		uri, exists := element.Attr("href")
		if exists {
			uris = append(uris, uri)
		}
	}
	return uris
}

func addHosturi(host string, uri string) string {
	if strings.Contains(uri, "http") {
		return uri
	} else {
		return host + uri
	}
}

func getType(uri string) (string, error) {
	for _, typ := range TypeOptions {
		if strings.Contains(uri, typ) {
			return typ, nil
		}
	}
	return "", fmt.Errorf("this card does not match any of types")
}

func fetchPage(root string, urlWithParams string, page int, ch chan []types.Offert) {
	docChannel := make(chan types.Doc)
	docErrChannel := make(chan error)
	offertChannel := make(chan types.Offert)
	offertErrChannel := make(chan error)

	uri := fmt.Sprintf("%s&page=%d", urlWithParams, page)
	doc := getDocOrPanic(uri)

	carduris := getCarduris(doc)
	cardsToHandle := 0
	offerts := []types.Offert{}

	for _, carduri := range carduris {
		carduriWithHost := addHosturi(root, carduri)
		typ, err := getType(carduriWithHost)
		if err != nil {
			continue
		}
		go getDoc(carduriWithHost, typ, docChannel, docErrChannel)
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
