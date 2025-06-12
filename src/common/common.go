package common

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func CleanNumberStr(str string) string {
	re := regexp.MustCompile(`[^\d\.]`)
	cleanStr := re.ReplaceAllString(str, "")
	return cleanStr
}

func ParsePrice(price string) (int, error) {
	digits := CleanNumberStr(price)
	result, err := strconv.Atoi(digits)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func getSelectionByTextContain(doc *goquery.Selection, text string) (*goquery.Selection, error) {
	items := doc.Find(`data-sentry-element="Item"`).EachIter()
	for _, item := range items {
		ci := item.Contents().EachIter()
		for _, content := range ci {
			if goquery.NodeName(content) != "#text" {
				continue
			}
			if strings.Contains(content.Text(), text) {
				return item, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to find element by text %s", text)
}

func GetSelectionBySiblingText(doc *goquery.Selection, text string) (*goquery.Selection, error) {
	sibling, err := getSelectionByTextContain(doc, text)
	if err != nil {
		return nil, err
	}
	siblings := sibling.Siblings()
	return siblings.First(), nil
}
