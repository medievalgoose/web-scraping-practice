package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type Pokemon struct {
	Name        string
	Price       string
	Description string
	Stock       int
}

func main() {
	pageIndex := 1

	// Create a new json file
	fileName := "pokemon.json"
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	// Close the file when main function is finished.
	defer file.Close()

	var pokemonList []Pokemon

	cl := colly.NewCollector(
		colly.CacheDir("./cache"),
	)

	detailCollector := cl.Clone()

	// Use either regular quote or backtick to specify which element to detect.
	// Context: check the HTML for the "li" element that has "type-product" class.
	cl.OnHTML(`li.type-product`, func(lmn *colly.HTMLElement) {
		// Get the AbsoluteURL from href attribute of the matching element specified in query selector.
		pokemonLink := lmn.Request.AbsoluteURL(lmn.ChildAttr("a.woocommerce-LoopProduct-link.woocommerce-loop-product__link", "href"))

		// Instruct detailCollector to visit the product description page.
		detailCollector.Visit(pokemonLink)
	})

	cl.OnScraped(func(r *colly.Response) {
		pageIndex++

		if pageIndex == 5 {
			return
		}

		// After the page is scraped, try to visit the next page.
		targetUrl := fmt.Sprintf("https://scrapeme.live/shop/page/%v", pageIndex)
		if err := r.Request.Visit(targetUrl); err != nil {
			return
		}
	})

	detailCollector.OnHTML(`div.summary.entry-summary`, func(h *colly.HTMLElement) {
		pokemonName := h.ChildText("h1.product_title.entry-title")
		pokemonPrice := h.ChildText("span.woocommerce-Price-amount.amount")
		pokemonDescription := h.ChildText("div.woocommerce-product-details__short-description > p")

		pokemonStockString := h.ChildText("p.stock")
		pokemonStockArray := strings.Split(pokemonStockString, " ")

		pokemonStock, err := strconv.Atoi(pokemonStockArray[0])
		if err != nil {
			pokemonStock = 0
		}

		// fmt.Printf("%v | %v | %v | %v\n", pokemonName, pokemonPrice, pokemonDescription, pokemonStock)

		newPokemon := Pokemon{
			Name:        pokemonName,
			Price:       pokemonPrice,
			Description: pokemonDescription,
			Stock:       pokemonStock,
		}

		pokemonList = append(pokemonList, newPokemon)
	})

	cl.Visit("https://scrapeme.live/shop/")

	// Encode the pokemonList into a json file.
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	enc.Encode(pokemonList)
}
