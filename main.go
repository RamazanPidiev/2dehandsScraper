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

type Product struct {
	URL   string
	Image string
	Title string
	Price string
}

func main() {

	var totalSum float64
	var totalProducts float64

	// Set up scraper and slow it down to reduce detection
	c := colly.NewCollector(
		colly.AllowedDomains("2dehands.be", "www.2dehands.be"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       5 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		// Set cookies if needed
		// r.Headers.Set("Cookie", "name=value")
		fmt.Println("Scraping", r.URL.String()+" ...")
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	var products []Product

	c.OnHTML("li.hz-Listing.hz-Listing--list-item", func(e *colly.HTMLElement) {
		if e.DOM.Find("span:contains('Topadvertentie'), span:contains('Topzoekertje') span:contains('Bezoek website')").Length() > 0 {
			return
		}
		href := e.ChildAttr("a.hz-Link.hz-Link--block.hz-Listing-coverLink", "href")

		html, err := e.DOM.Html()

		if err == nil {

			fmt.Println("HTML:", html)
			fmt.Println()
			fmt.Println()
			fmt.Println()
			fmt.Println("---------------------------------------------------")
		}

		product := Product{
			URL:   "https://www.2dehands.be" + href,
			Image: e.ChildAttr("img", "src"),
			Title: e.ChildText("h3.hz-Listing-title"),
			Price: e.ChildText("div.hz-Listing-price-extended-details p"),
		}

		if product.Price == "Bieden" || product.Price == "Ruilen" {
			return
		}

		priceFloat, err2 := parsePrice(product.Price)

		if err2 != nil {
			fmt.Println("Error:", err2)
			return
		}

		if priceFloat/100 < 400 {
			fmt.Println("Product removed: ", product.Title, " ", priceFloat/100, " EURO")
			return
		}

		totalSum += priceFloat
		totalProducts++

		products = append(products, product)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

	for i := 1; i <= 2; i++ {
		query := "iphone+15"
		headers := "#Language:all-languages|searchInTitleAndDescription:true"
		url := fmt.Sprintf("https://www.2dehands.be/q/%s/p/%d/%s", query, i, headers)
		c.Visit(url)
	}

	// Open CSV file to append data
	file, err := os.OpenFile("products.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open output CSV file", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Check if file is empty to write headers
	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Failed to get file info", err)
		return
	}

	if stat.Size() == 0 {
		headers := []string{
			"URL",
			"Image",
			"Name",
			"Price",
		}
		writer.Write(headers)
	}

	for _, product := range products {
		record := []string{
			product.URL,
			product.Image,
			product.Title,
			strings.TrimSpace(product.Price),
		}
		writer.Write(record)
	}

	fmt.Println("Average price: ", ((totalSum / 100) / totalProducts))
}

func parsePrice(priceStr string) (float64, error) {
	cleanedStr := strings.ReplaceAll(priceStr, "€", "")
	cleanedStr = strings.ReplaceAll(cleanedStr, ",", ".")
	cleanedStr = strings.ReplaceAll(cleanedStr, ".", "")
	cleanedStr = strings.TrimSpace(cleanedStr)
	return strconv.ParseFloat(cleanedStr, 64)
}
