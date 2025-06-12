package olx

import (
	"errors"
	"regexp"
	"strings"

	"house-pricer/common"
	"house-pricer/types"
)

const OffertType = "olx"

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
