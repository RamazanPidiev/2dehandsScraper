package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Product2 struct {
	URL   string
	Image string
	Title string
	Price string
}

func main2() {
	// set up scraper and slow it down to reduce detection
	c := colly.NewCollector(
		colly.AllowedDomains("2dehands.be", "www.2dehands.be"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       20 * time.Second,
	})

	// This will get called for each request
	c.OnRequest(func(r *colly.Request) {
		// Set cookies if needed
		// r.Headers.Set("Cookie", "name=value")
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Product slice
	var products []Product

	c.OnHTML("div.hz-Listing-listview-content", func(e *colly.HTMLElement) {
		if e.DOM.Find("span:contains('Topadvertentie'), span:contains('Topzoekertje')").Length() > 0 {
			return
		}

		product := Product{
			URL:   "https://www.2dehands.be" + e.ChildAttr("a", "href"),
			Image: e.ChildAttr("img", "src"),
			Title: e.ChildText("h3.hz-Listing-title"),
			Price: e.ChildText("div.hz-Listing-price-extended-details p"),
		}

        if product.Price == "Bieden" {
			return
		}

		cleanedStr := strings.ReplaceAll(product.Price, "€", "")
		cleanedStr = strings.ReplaceAll(cleanedStr, ",", ".")
		cleanedStr = strings.ReplaceAll(cleanedStr, ".", "")
		cleanedStr = strings.TrimSpace(cleanedStr)

		priceFloat, err := strconv.ParseFloat(cleanedStr, 64)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if (priceFloat / 100 < 400) { 
            return
        }

		products = append(products, product)
	})

	c.Visit("https://www.2dehands.be/q/iphone+15")

	// Create CSV file to insert data
	file, err := os.Create("products.csv")
	if err != nil {
		fmt.Println("Failed to create output CSV file", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"URL",
		"Image",
		"Name",
		"Price",
	}

	writer.Write(headers)

	// Write data to CSV
	for _, product := range products {
		record := []string{
			product.URL,
			product.Image,
			product.Title,
			strings.TrimSpace(product.Price),
		}
		writer.Write(record)
	}
}
