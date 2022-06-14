package blogtruyen

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"
)

var (
	host      = "https://blogtruyen.vn"
	searchFmt = "/ajax/Search/AjaxLoadListManga?key=tatca&order=%v&p=%v"
)

type MangaOrder uint

const (
	Default MangaOrder = iota
	Name
	Chapter
	View
	Comment
	Time
)

func BuildURL(segments ...string) string {
	return fmt.Sprintf("%s%s", host, strings.Join(segments, "/"))
}

func ListMangaURL(order MangaOrder, page uint) string {
	segments := fmt.Sprintf(searchFmt, order, page)
	return BuildURL(segments)
}

func GET(url string) (doc *goquery.Document) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func Parse(doc *goquery.Document, selector string, attr string) (res []string) {
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		var data string
		var exist bool
		if attr == "" || attr == "text" {
			data = s.Text()
		} else {
			data, exist = s.Attr(attr)
			if !exist {
				return
			}
		}
		res = append(res, data)
	})
	return
}

func ParseMangaLinks(doc *goquery.Document) (mangaLinks []string) {
	mangaLinks = Parse(doc, ".list > p > span > a", "href")
	mangaLinks = lo.Map[string, string](mangaLinks, func(link string, _ int) string {
		return BuildURL(link)
	})
	return
}

func ParseLastPageNumber(doc *goquery.Document) uint {
	res := Parse(doc, "span.page:last-child > a", "href")
	if len(res) == 0 {
		return 0
	}
	numStr := strings.FieldsFunc(res[0], func(r rune) bool {
		return r == '(' || r == ')'
	})[1]
	lastPage, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return uint(lastPage)
}

func ExampleScrape() {
	order := Time
	page := uint(1)
	url := ListMangaURL(order, page)
	doc := GET(url)
	links := ParseMangaLinks(doc)
	fmt.Println("Links:")
	for _, link := range links {
		fmt.Println(link)
	}
	lastPage := ParseLastPageNumber(doc)
	fmt.Printf("Last page: %v\n", lastPage)
}
