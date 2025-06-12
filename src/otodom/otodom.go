package otodom

import (
	"errors"
	"strings"

	"house-pricer/common"
	"house-pricer/types"
)

const OffertType = "otodom"

func getFurnishedVal(str string) int {
	if strings.Contains(str, "do zamieszkania") {
		return 1
	} else if strings.Contains(str, "do wykończenia") {
		return -1
	} else {
		return 0
	}
}

func GetOffert(doc types.Doc, offertChannel chan types.Offert, offertErrChannel chan error) {
	fail := func(err error) bool {
		if err != nil {
			offertErrChannel <- errors.New(doc.Url + " " + err.Error())
			return true
		}
		return false
	}

	priceStr := doc.Doc.Find(`[aria-label="Cena"]`).First().Text()
	price, err := common.ParsePrice(priceStr)
	if fail(err) {
		return
	}

	title := doc.Doc.Find(`[data-cy="adPageAdTitle"]`).First().Text()
	container := doc.Doc.Find(`[data-sentry-element="StyledListContainer"]`).First()

	areaEl, err := common.GetSelectionBySiblingText(container, "Powierzchnia")
	if fail(err) {
		return
	}
	area := common.CleanNumberStr(areaEl.Text())

	furnishedEl, err := common.GetSelectionBySiblingText(container, "Stan wykończenia")
	if fail(err) {
		return
	}
	furnished := getFurnishedVal(furnishedEl.Text())

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
