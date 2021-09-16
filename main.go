package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/gocolly/colly"
)

type Product struct {
	name        string
	description string
	dimension   string
	features    string
	price       float64
}

type CollectionCategory struct {
	name     string
	products []*Product
}
type Collection struct {
	name       string
	categories []*CollectionCategory
}
type Page struct {
	name        string
	collections []*Collection
}

func getLastSubPath(url string) (string, error) {
	validUrl := regexp.MustCompile(`(\S+/)+(\S+)/*`)
	m := validUrl.FindStringSubmatch(url)
	if m != nil {
		return m[len(m)-1], nil
	} else {
		return "", fmt.Errorf("Invalid url")
	}
}

const (
	homePage                       = "https://interiorsonline.com.au"
	pageLinkTemplate               = homePage + "/pages/%v"
	collectionLinkTemplate         = pageLinkTemplate + "/collections/%v"
	collectionCategoryLinkTemplate = collectionLinkTemplate + "/category-%v"
)

var pageChann = make(chan *Page)
var collectionChann = make(chan *Collection)
var collectionCategoryChann = make(chan *CollectionCategory)
var productChann = make(chan *Product)

var pageTable = make([]*Page, 0)
var collectionTable = make([]*Collection, 0)
var collectionCategoryTable = make([]*CollectionCategory, 0)
var productTable = make([]*Product, 0)

func main() {
	collector1 := colly.NewCollector(colly.Async(true))
	collector2 := colly.NewCollector(colly.Async(true))
	collector1.OnHTML("ul#main-nav", func(h1 *colly.HTMLElement) {
		h1.ForEach("li.dropdown >a", func(i int, h2 *colly.HTMLElement) {
			if i < 4 {
				pageLink := h2.Request.AbsoluteURL(h2.Attr("href"))
				name, err := getLastSubPath(h2.Request.AbsoluteURL(pageLink))
				if err != nil {
					log.Printf("%q is invalid url", pageLink)
				} else {
					newPage := Page{
						name: name,
					}
					pageChann <- &newPage
					pageTable = append(pageTable, &newPage)
				}
			}
		})

	})
	collector1.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	collector2.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	collector1.Visit(homePage)

	for page := range pageChann {
		linkToVisit := fmt.Sprintf(pageLinkTemplate, page.name)
		collector2.Visit(linkToVisit)

	}
	collector1.Wait()
	collector2.Wait()

}
