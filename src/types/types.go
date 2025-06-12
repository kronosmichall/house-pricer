package types

import (
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Offert struct {
	URL       string  `json:"url" dynamodbav:"URL"` // Partition Key, string type
	Title     string  `json:"title" dynamodbav:"Title"`
	Type      string  `json:"type" dynamodbav:"Type"`
	Price     int     `json:"price" dynamodbav:"Price"`
	Area      string  `json:"area" dynamodbav:"Area"`
	Mortgage  float64 `json:"mortgage" dynamodbav:"Mortgage"`
	Furnished int     `json:"furnished" dynamodbav:"Furnished"` // 1 true, 0 unknown, -1 false
	// Added timestamps for better record keeping.
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"CreatedAt"`     // Stored as ISO 8601 string
	LastUpdated time.Time `json:"lastUpdated" dynamodbav:"LastUpdated"` // Stored as ISO 8601 string
}

type Doc struct {
	Doc  *goquery.Document
	Url  string
	Type string
}
