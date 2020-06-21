package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Utils struct {}

func (u *Utils) MysqlRealEscapeString(value string) string {
	replace := map[string]string{"\\":"\\\\", "'":`\'`, "\\0":"\\\\0", "\n":"\\n", "\r":"\\r", `"`:`\"`, "\x1a":"\\Z"}

	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}

	return value;
}

func (u *Utils) SetInterval(someFunc func(), milliseconds int, async bool) chan bool {

	// How often to fire the passed in function
	// in milliseconds
	interval := time.Duration(milliseconds) * time.Millisecond

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(interval)
	clear := make(chan bool)

	// Put the selection in a go routine
	// so that the for loop is none blocking
	go func() {
		for {

			select {
			case <-ticker.C:
				if async {
					// This won't block
					go someFunc()
				} else {
					// This will block
					someFunc()
				}
			case <-clear:
				ticker.Stop()
				return
			}

		}
	}()

	// We return the channel so we can pass in
	// a value to it to clear the interval
	return clear
}

func (u *Utils) ErrorHandler(err error) {
	if err != nil {
		log.Println(err)
	}
}

func (u *Utils) SentenceSplit(str string) []string {
	return u.PregSplit(str, `([.?!])\s+`)
}

func (u *Utils) PregSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes) + 1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:len(text)]
	return result
}

func (u *Utils) YoutubeEmbed(str string) string {
	if strings.Contains(str, "youtube.com/watch?v=") {
		str = strings.Replace(str, "youtube.com/watch?v=", "youtube.com/embed/", 1)
	} else if strings.Contains(str, "youtu.be/") {
		str = strings.Replace(str, "youtu.be/", "youtube.com/embed/", 1)
	} else if strings.Contains(str, "/watch?v=") {
		str = strings.Replace(str, "/watch?v=", "youtube.com/embed/", 1)
	}
	str = strings.Replace(str, "&", "?", 1)
	return `https://` + str
}


func (u *Utils) Format(str string) string {
	reg := regexp.MustCompile(`(?m)<div[^<>]*><a[^<>]*>More items...</a></div>`)
	str = reg.ReplaceAllString(str, ``)

	reg = regexp.MustCompile(`(?m)<div[^<>]*>•</div>`)
	str = reg.ReplaceAllString(str, ``)

	reg = regexp.MustCompile(`(?m)<div[^<>]*>[JFMASOND][a-z]{2}\s\d{1,2},\s\d{4}</div>`)
	str = reg.ReplaceAllString(str, ``)

	reg = regexp.MustCompile(`(?m)<span[^<>]*>[JFMASOND][a-z]{2}\s\d{1,2},\s\d{4}</span>`)
	str = reg.ReplaceAllString(str, ``)

	headingMatch := utils.PregMatch(`(?m)<div[^<>]*role="heading"><b>(?P<title>.+)</b></div>`, str)
	heading := headingMatch["title"]
	heading = utils.StripTags(heading)

	str, _ = sanitize.HTMLAllowing(str, []string{
		"table", "thead", "tbody", "tr", "td", "th",
		"ol", "ul", "li",
		"dl", "dt", "dd",
		"p", "br"})

	str = strings.Replace(str, "...", ".", -1)

	str = strings.Replace(str, heading, "<h3>" + heading + "</h3>", 1)

	return str
}

func (u *Utils) StripTags(html string) string {
	paaReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(paaReader)
	if err != nil {
		log.Println(err)
	}

	return doc.Text()
}

func (u *Utils) RandBool() bool {
	return rand.Float32() < 0.5
}

func (u *Utils) ParseFormCollection(r *http.Request, typeName string) map[string]string {
	result := make(map[string]string)
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	for key, values := range r.Form {
		re := regexp.MustCompile(typeName + "\\[(.+)\\]")
		matches := re.FindStringSubmatch(key)

		if len(matches) >= 2 {
			result[matches[1]] = values[0]
		}
	}
	return result
}

func (u *Utils) toInt(value string) int {
	var integer int = 0
	if value != "" {
		integer, _ = strconv.Atoi(value)
	}
	return integer
}

func (u *Utils) PregMatch(regEx, url string) (paramsMap map[string]string) {
	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return
}


// Counters - work with mutex
type Counters struct {
	mx sync.Mutex
	m map[string]int
}

func (c *Counters) Load(key string) (int, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	val, ok := c.m[key]
	return val, ok
}

func (c *Counters) Store(key string, value int) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m[key] = value
}

func NewCounters() *Counters {
	return &Counters{
		m: make(map[string]int),
	}
}