package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/gocolly/colly"
)

type Product struct {
	name               string
	description        []string
	featureList        []string
	collectionCategory string
}

type CollectionCategory struct {
	name       string
	collection string
	link       string
}
type Collection struct {
	name string
	page string
	link string
}
type Page struct {
	name string
	link string
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

var pageTable = make([]*Page, 0)
var collectionTable = make([]*Collection, 0)
var collectionCategoryTable = make([]*CollectionCategory, 0)
var productTable = make([]*Product, 0)

func main() {
	pageCollector := colly.NewCollector(colly.Async(true))
	collectionCollector := colly.NewCollector(colly.Async(true))
	collectionCategoryCollector := colly.NewCollector(colly.Async(true))
	productCollector := colly.NewCollector(colly.Async(true))
	productDetailCollector := colly.NewCollector(colly.Async(true))
	pageCollector.OnHTML("ul#main-nav", func(h1 *colly.HTMLElement) {
		h1.ForEach("li.dropdown >a", func(i int, h2 *colly.HTMLElement) {
			if i < 4 {
				pageLink := h2.Request.AbsoluteURL(h2.Attr("href"))
				name, err := getLastSubPath(h2.Request.AbsoluteURL(pageLink))
				if err != nil {
					log.Printf("%q is invalid url", pageLink)
				} else {
					newPage := Page{
						name: name,
						link: pageLink,
					}
					pageChann <- &newPage
					pageTable = append(pageTable, &newPage)
				}
			}
		})
		close(pageChann)

	})
	collectionCollector.OnHTML("body", func(h *colly.HTMLElement) {
		h.ForEach("div.collection-info-inner > a", func(i int, h1 *colly.HTMLElement) {
			collectionLink := h1.Request.AbsoluteURL(h1.Attr("href"))
			name, err := getLastSubPath(h1.Request.AbsoluteURL(collectionLink))
			if err != nil {
				log.Printf("%q is invalid url", collectionLink)
			} else {
				newCollection := Collection{
					name: name,
					link: collectionLink,
					page: h1.Request.Ctx.Get("owner"),
				}
				collectionChann <- &newCollection
				collectionTable = append(collectionTable, &newCollection)
			}
		})
		close(collectionChann)

	})
	collectionCategoryCollector.OnHTML("select.styled-select.coll-filter.drop-a", func(h *colly.HTMLElement) {
		h.ForEach("option", func(i int, h1 *colly.HTMLElement) {
			collectionCategoryLink := h1.Request.AbsoluteURL(h1.Attr("value"))
			name, err := getLastSubPath(h1.Request.AbsoluteURL(collectionCategoryLink))
			if err != nil {
				log.Printf("%q is invalid url", collectionCategoryLink)
			} else {
				newCollectionCategory := CollectionCategory{
					name:       name,
					link:       collectionCategoryLink,
					collection: h1.Request.Ctx.Get("owner"),
				}
				collectionCategoryChann <- &newCollectionCategory
				collectionCategoryTable = append(collectionCategoryTable, &newCollectionCategory)
			}

		})
		close(collectionCategoryChann)

	})
	productCollector.OnHTML(".prod-image", func(h *colly.HTMLElement) {
		h.ForEach("a", func(i int, h1 *colly.HTMLElement) {
			productLink := h1.Request.AbsoluteURL(h1.Attr("href"))
			newContext := colly.NewContext()
			newContext.Put("owner", h1.Request.Ctx.Get("owner"))
			productDetailCollector.Request("GET", productLink, nil, newContext, nil)
		})

	})
	productDetailCollector.OnHTML("body", func(h *colly.HTMLElement) {
		newProduct := Product{
			collectionCategory: h.Request.Ctx.Get("owner"),
		}
		newProduct.name = h.ChildText("h1")
		h.ForEach(".lower-description1>p", func(i int, h1 *colly.HTMLElement) {
			newProduct.description = append(newProduct.description, h1.Text)
		})
		h.ForEach(".lower-description1>ul>li", func(i int, h2 *colly.HTMLElement) {
			newProduct.featureList = append(newProduct.featureList, h2.Text)
		})
		productTable = append(productTable, &newProduct)
	})
	pageCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	collectionCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	pageCollector.Visit(homePage)

	for page := range pageChann {
		linkToVisit := fmt.Sprintf(pageLinkTemplate, page.name)
		ctx := colly.NewContext()
		ctx.Put("owner", page.name)
		collectionCollector.Request("GET", linkToVisit, nil, ctx, nil)
	}
	for collection := range collectionChann {
		linkToVisit := collection.link
		ctx := colly.NewContext()
		ctx.Put("owner", collection.name)
		collectionCategoryCollector.Request("GET", linkToVisit, nil, ctx, nil)
	}
	for collectionCategory := range collectionCategoryChann {
		linkToVisit := collectionCategory.link
		ctx := colly.NewContext()
		ctx.Put("owner", collectionCategory.name)
		productCollector.Request("GET", linkToVisit, nil, ctx, nil)
	}

}
