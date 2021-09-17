package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"example.com/crawler/tocsv"
	"github.com/gocolly/colly"
)

type Product struct {
	Name               string
	Description        string
	FeatureList        string
	Price              float64
	Image              string
	CollectionCategory string
	Collection         string
}

type Category struct {
	Name       string
	Collection string
}
type Collection struct {
	Name  string
	Page  string
	Image string
}
type Page struct {
	Name string
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

var validPriceString = regexp.MustCompile(`\$(\d+),?(\d*\.\d+)`)
var (
	pageTable       = make([]*Page, 0)
	collectionTable = make([]*Collection, 0)
	categoryTable   = make([]*Category, 0)
	productTable    = make([]*Product, 0)
)

func main() {
	pageCollector := colly.NewCollector(colly.Async(true))
	collectionCollector := colly.NewCollector(colly.Async(true))
	categoryCollector := colly.NewCollector(colly.Async(true))
	productCollector := colly.NewCollector(colly.Async(true))
	productDetailCollector := colly.NewCollector(colly.Async(true))

	pageCollector.OnHTML("ul#main-nav", func(h1 *colly.HTMLElement) {
		log.Println("found you")
		h1.ForEach("li.dropdown >a", func(i int, h2 *colly.HTMLElement) {
			if i < 4 {
				pageLink := h2.Request.AbsoluteURL(h2.Attr("href"))
				name, err := getLastSubPath(h2.Request.AbsoluteURL(pageLink))
				log.Println(pageLink)
				if err != nil {
					log.Printf("%q is invalid url", pageLink)
				} else {
					log.Println("found new page")
					newPage := Page{
						Name: name,
					}
					pageTable = append(pageTable, &newPage)
					newContext := colly.NewContext()
					newContext.Put("page", name)
					collectionCollector.Request("GET", pageLink, nil, newContext, nil)
				}
			}
		})

	})
	collectionCollector.OnHTML("body", func(h *colly.HTMLElement) {
		h.ForEach("div.collection-image", func(i int, h1 *colly.HTMLElement) {
			collectionLink := h1.Request.AbsoluteURL(h1.ChildAttr("a", "href"))
			name, err := getLastSubPath(h1.Request.AbsoluteURL(collectionLink))
			if err != nil {
				log.Printf("%q is invalid url", collectionLink)
			} else {
				newCollection := Collection{
					Name:  name,
					Page:  h1.Request.Ctx.Get("page"),
					Image: h1.ChildAttr("img", "src"),
				}
				collectionTable = append(collectionTable, &newCollection)
				newContext := colly.NewContext()
				newContext.Put("collection", name)
				categoryCollector.Request("GET", collectionLink, nil, newContext, nil)
			}
		})

	})
	categoryCollector.OnHTML("select.styled-select.coll-filter.drop-a", func(h *colly.HTMLElement) {
		h.ForEach("option", func(i int, h1 *colly.HTMLElement) {
			name := h1.Attr("value")
			if name != "" {
				collectionCategoryLink := h.Request.URL.String() + "/" + name
				newCategory := Category{
					Name:       name,
					Collection: h1.Request.Ctx.Get("collection"),
				}
				categoryTable = append(categoryTable, &newCategory)
				newContext := colly.NewContext()
				newContext.Put("category", name)
				newContext.Put("collection", h1.Request.Ctx.Get("collection"))
				productCollector.Request("GET", collectionCategoryLink, nil, newContext, nil)
			}
		})

	})
	productCollector.OnHTML(".prod-image", func(h *colly.HTMLElement) {
		h.ForEach("a", func(i int, h1 *colly.HTMLElement) {
			productLink := h1.Request.AbsoluteURL(h1.Attr("href"))
			productImage := h1.ChildAttr("a>img", "src")
			newContext := colly.NewContext()
			newContext.Put("category", h1.Request.Ctx.Get("category"))
			newContext.Put("collection", h1.Request.Ctx.Get("collection"))
			newContext.Put("image", productImage)
			productDetailCollector.Request("GET", productLink, nil, newContext, nil)
		})

	})
	productDetailCollector.OnHTML("body", func(h *colly.HTMLElement) {
		newProduct := Product{
			CollectionCategory: h.Request.Ctx.Get("category"),
			Collection:         h.Request.Ctx.Get("collection"),
			Image:              h.Request.Ctx.Get("image"),
			Description:        "",
			FeatureList:        "",
		}
		newProduct.Name = h.ChildText("h1")
		priceString := h.ChildText("span.product-price")
		var price float64
		var err error
		m := validPriceString.FindStringSubmatch(priceString)
		log.Println(m)
		if m != nil {
			price, err = strconv.ParseFloat(strings.Join(m[1:], ""), 64)
			if err != nil {
				log.Println("err extracting price", err)
			}
		}
		log.Println(price)

		newProduct.Price = price
		h.ForEach("#lower-description1>p", func(i int, h1 *colly.HTMLElement) {
			newProduct.Description += "||" + h1.Text
		})
		h.ForEach("#lower-description1>ul>li", func(i int, h2 *colly.HTMLElement) {
			newProduct.FeatureList += "||" + h2.Text
		})
		productTable = append(productTable, &newProduct)
	})
	pageCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	collectionCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	categoryCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	productCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	productDetailCollector.OnRequest(func(r *colly.Request) { log.Printf("Visiting %q", r.URL) })
	pageCollector.Visit(homePage)
	pageCollector.Wait()
	collectionCollector.Wait()
	categoryCollector.Wait()
	productCollector.Wait()
	productDetailCollector.Wait()

	tocsv.WriteCsv(tocsv.ObjSlice2SliceSlice(pageTable), "page.csv")
	tocsv.WriteCsv(tocsv.ObjSlice2SliceSlice(collectionTable), "collection.csv")
	tocsv.WriteCsv(tocsv.ObjSlice2SliceSlice(categoryTable), "category.csv")
	tocsv.WriteCsv(tocsv.ObjSlice2SliceSlice(productTable), "product.csv")

}
