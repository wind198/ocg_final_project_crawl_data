package main

import (
	"fmt"
	"log"
	"regexp"

	"example/craw-db-interior-ecommerce-site/tocsv"

	"github.com/gocolly/colly"
)

type Product struct {
	name               string
	description        string
	featureList        string
	collectionCategory string
}

type CollectionCategory struct {
	name       string
	collection string
}
type Collection struct {
	name string
	page string
}
type Page struct {
	name string
}

func (p *Page) toSlice([]interface{}) {

}

func getLastSubPath(url string) (string, error) {
	validUrl := regexp.MustCompile(`(\S+/)+(\S+)/*`)
	m := validUrl.FindStringSubmatch(url)
	if m != nil {
		return m[len(m)-1], nil
	} else {
		return "", fmt.Errorf("INVALID URL")
	}
}

const (
	homePage                       = "https://interiorsonline.com.au"
	pageLinkTemplate               = homePage + "/pages/%v"
	collectionLinkTemplate         = pageLinkTemplate + "/collections/%v"
	collectionCategoryLinkTemplate = collectionLinkTemplate + "/category-%v"
)

var (
	pageTable               = make([]*Page, 0)
	collectionTable         = make([]*Collection, 0)
	collectionCategoryTable = make([]*CollectionCategory, 0)
	productTable            = make([]*Product, 0)
)
var (
	productCSVdata            = make([][]string, 0)
	collectioinCSVdata        = make([][]string, 0)
	collectionCategoryCSVdata = make([][]string, 0)
	pageCSVdata               = make([][]string, 0)
)

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
					}
					pageTable = append(pageTable, &newPage)
					newContext := colly.NewContext()
					newContext.Put("owner", name)
					collectionCollector.Request("GET", pageLink, nil, newContext, nil)
				}
			}
		})

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
					page: h1.Request.Ctx.Get("owner"),
				}
				collectionTable = append(collectionTable, &newCollection)
				newContext := colly.NewContext()
				newContext.Put("owner", name)
				collectionCategoryCollector.Request("GET", collectionLink, nil, newContext, nil)
			}
		})

	})
	collectionCategoryCollector.OnHTML("select.styled-select.coll-filter.drop-a", func(h *colly.HTMLElement) {
		h.ForEach("option", func(i int, h1 *colly.HTMLElement) {
			name := h1.Attr("value")
			if name != "" {
				collectionCategoryLink := h.Request.URL.String() + "/" + name
				newCollectionCategory := CollectionCategory{
					name:       name,
					collection: h1.Request.Ctx.Get("owner"),
				}
				collectionCategoryTable = append(collectionCategoryTable, &newCollectionCategory)
				newContext := colly.NewContext()
				newContext.Put("owner", name)
				productCollector.Request("GET", collectionCategoryLink, nil, newContext, nil)
			}
		})

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
			description:        "",
			featureList:        "",
		}
		newProduct.name = h.ChildText("h1")
		h.ForEach("#lower-description1>p", func(i int, h1 *colly.HTMLElement) {
			newProduct.description += "||" + h1.Text
		})
		h.ForEach("#lower-description1>ul>li", func(i int, h2 *colly.HTMLElement) {
			newProduct.featureList += "||" + h2.Text
		})
		productTable = append(productTable, &newProduct)
	})
	// pageCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	// collectionCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	// collectionCategoryCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	// productCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	// productDetailCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	pageCollector.Visit(homePage)
	pageCollector.Wait()
	collectionCollector.Wait()
	collectionCategoryCollector.Wait()
	productCollector.Wait()
	productDetailCollector.Wait()

	tocsv.WriteCsv(tocsv.ObjSlice2SliceSlice(pageTable), "pages")

}
