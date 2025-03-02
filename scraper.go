package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gocolly/colly"
)

// Product structure to store scraped data
type Product struct {
	Name, Price, URL, Image string
}

func main() {
	// Initialize Colly collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.scrapingcourse.com"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		Parallelism: 4,
	})

	var products []Product
	var visitedUrls sync.Map

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

	// OnHTML callback for scraping product information
	c.OnHTML("li.product", func(e *colly.HTMLElement) {
		product := Product{
			Name:  e.ChildText("h2.woocommerce-loop-product__title"),
			Price: e.ChildText("span.price"),
			URL:   e.ChildAttr("a.woocommerce-LoopProduct-link", "href"),
			Image: e.ChildAttr("img.attachment-woocommerce_thumbnail", "src"),
		}
		products = append(products, product)
	})

	// Handle pagination
	c.OnHTML("a.next", func(e *colly.HTMLElement) {
		nextPage := e.Attr("href")
		if _, found := visitedUrls.Load(nextPage); !found {
			fmt.Println("Scraping next page:", nextPage)
			visitedUrls.Store(nextPage, struct{}{})
			e.Request.Visit(nextPage)
		}
	})

	// Store the data to a CSV file after scraping
	c.OnScraped(func(r *colly.Response) {
		file, err := os.Create("products.csv")
		if err != nil {
			log.Fatalln("Failed to create output CSV file", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		writer.Write([]string{"Name", "Price", "URL", "Image"})

		for _, product := range products {
			writer.Write([]string{product.Name, product.Price, product.URL, product.Image})
		}
		writer.Flush()
	})

	// Start scraping
	c.Visit("https://www.scrapingcourse.com/ecommerce/")

	c.Wait()
}
