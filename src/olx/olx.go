package olx

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"house-pricer/common"
	"house-pricer/types"
)

const (
	Root = "https://www.olx.pl"
	Home = "https://www.olx.pl/nieruchomosci/mieszkania/sprzedaz/warszawa"
)
const OffertType = "olx"

var PriceOptions = [...]int{
	140_000,
	160_000,
	180_000,
	200_000,
	250_000,
	300_000,
	350_000,
	400_000,
	500_000,
	600_000,
	700_000,
	800_000,
	1_000_000,
}

func UrlWithParamsGenerator() <-chan string {
	uri, err := url.Parse(Home)
	if err != nil {
		panic(err)
	}
	q := uri.Query()

	out := make(chan string)

	go func() {
		defer close(out)
		for i := range len(PriceOptions) {
			q.Del("search[filter_float_price:from]")
			q.Del("search[filter_float_price:to]")

			if i > 0 {
				price := strconv.Itoa(PriceOptions[i])
				q.Add("search[filter_float_price:from]", price)
			}
			if i < len(PriceOptions)-1 {
				price := strconv.Itoa(PriceOptions[i+1])
				q.Add("search[filter_float_price:to]", price)
			}

			uri.RawQuery = q.Encode()
			str := uri.String()
			str = strings.Replace(str, "?", "/?", 1)
			str = strings.ReplaceAll(str, "%3A", ":")

			out <- str
		}
	}()
	return out
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
		return "?", errors.New("no area tag")
	}
	areaStr := matches[1]
	return common.CleanNumberStr(areaStr), nil
}

func GetOffert(doc types.Doc, offertChannel chan types.Offert, offertErrChannel chan error) {
	element := doc.Doc.Find(`div[data-testid="ad-price-container"] h3`).First()
	priceStr := element.Text()
	price, err := common.ParsePrice(priceStr)
	if err != nil {
		offertErrChannel <- errors.New(doc.Url + " " + err.Error())
		return
	}
	element = doc.Doc.Find(`[data-cy="offer_title"] h4`).First()
	title := element.Text()

	tagContainer := doc.Doc.Find(`div[data-testid="ad-parameters-container"]`).First()
	text := tagContainer.Text()
	furnished := getFurnished(text)
	area, err := getArea(text)
	if err != nil {
		offertErrChannel <- errors.New(doc.Url + " " + err.Error())
	}

	offert := types.Offert{
		URL:       doc.Url,
		Type:      OffertType,
		Title:     title,
		Price:     price,
		Area:      area,
		Mortgage:  0,
		Furnished: furnished,
	}
	offertChannel <- offert
}
