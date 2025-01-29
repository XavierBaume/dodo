package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// client create a new http client
func client() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
	}
}

// scrapDocs init the scrapping process with the current page we want to start with
func scrapDocs(currentPage int, language string) {
	_ = prepareDirs()
	wg := sync.WaitGroup{}
	// A general purpose scrapper with 8 instances and a sequential ticker
	ticker := time.NewTicker(400 * time.Millisecond)
	msg := make(chan message)
	urls := make(chan string, 10)
	for i := 0; i < 8; i++ {
		go scrap(urls, msg, ticker.C)
	}
	// handle documents after scrapping
	sitemapHandler := newSitemap(currentPage, language)
	go func() {
		wg.Add(1)
		defer wg.Done()
		for m := range msg {
			switch m.Type {
			case JSON:
				go func() {
					err := postDoc(urls, m)
					if err != nil {
						log.Println("Can't handle document", err)
					}
				}()
			case SITEMAP:
				go func() {
					err := sitemapHandler.postSitemap(urls, m)
					if err != nil {
						log.Println("Can't handle sitemap, try next", err)
					}
				}()
			case HTML:
			case FACSIMILE:
			}
		}
	}()
	// Wait until the goroutine is running and the wg is blocking
	time.Sleep(1 * time.Millisecond)
	// Process the sitemap on page XXX
	urls <- sitemapHandler.sitemapURL + "&p=" + strconv.Itoa(currentPage)
	log.Println("Wait for data")
	wg.Wait()
	log.Println("We are done.")
}

// prepareDirs prepare dirs to store dodis docs
func prepareDirs() error {
	err := os.MkdirAll("docs", 0755)
	if err != nil {
		return err
	}
	return nil
}

// scrap a single url and send the data to the msg chanel
func scrap(urls chan string, messages chan message, ticker <-chan time.Time) {
	for u := range urls {
		// block for the ticker. We want to send a request all 100ms
		<-ticker
		log.Println("Scrap data from ", u)
		// make request obj
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
		if !strings.Contains(u, "xml") {
			req.Header.Set("Content-Type", "application/json")
		}
		res, err := client().Do(req)
		if err != nil {
			log.Printf("Error scrapping %s, %v", u, err)
			cancel()
			continue
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			cancel()
			log.Printf("Error reading  %s into memory, %v", u, err)
			continue
		}
		url := parse(u)
		err = saveFile(url.Filename, data)
		if err != nil {
			cancel()
			log.Printf("Error writing  %s to dik, %v", u, err)
			continue
		}
		cancel()
		messages <- message{
			URL:     u,
			Content: data,
			Type:    url.Type,
		}
	}
}

// URLClassifier helps to guess a filepath and a type from an url
type URLClassifier struct {
	Type     documentType
	Filename string
}

// parse tages an url string and guesses the type and the filepath from it
func parse(url string) *URLClassifier {
	u := &URLClassifier{}
	if strings.Contains(url, "search") {
		u.Type = SITEMAP
		u.Filename = "sitemap"
		re := regexp.MustCompile(`p=(\d+)`)
		pages := re.FindAllStringSubmatch(url, 1)
		if len(pages) > 0 && len(pages[0]) > 1 {
			u.Filename += "-" + pages[0][1] + ".json"
		} else {
			u.Filename += "-0.json"
		}
	} else if strings.Contains(url, "html") {
		u.Type = HTML
		id := getID(url)
		if len(id) > 0 && len(id[0]) > 1 {
			u.Filename += id[0][1] + ".html"
		}
	} else if strings.Contains(url, "pdf") {
		u.Type = FACSIMILE
		id := getID(url)
		if len(id) > 0 && len(id[0]) > 1 {
			u.Filename += id[0][1] + ".pdf"
		}
	} else {
		u.Type = JSON
		re := regexp.MustCompile(`(\d+)`)
		id := re.FindAllStringSubmatch(url, 1)
		if len(id) > 0 && len(id[0]) > 1 {
			u.Filename += id[0][1] + ".json"
		}
	}
	return u
}

// get the id from an url
func getID(url string) [][]string {
	re := regexp.MustCompile(`dodis-(\d+)`)
	return re.FindAllStringSubmatch(url, 1)
}

// saveFile saves a file to disk
func saveFile(filename string, data []byte) error {
	if filename == "" {
		return errors.New("no valid filename")
	}
	err := os.WriteFile(path.Join("docs", filename), data, 0777)
	if err != nil {
		return err
	}
	return nil
}

// postDoc will handle all downloaded documents
// It will save the json on disk
// It will extract all signed xml links and queue those lnks
func postDoc(urls chan string, m message) error {
	// log.Println("Process document data ")
	// find all xml links and download them
	var d doc
	err := json.Unmarshal(m.Content, &d)
	if err != nil {
		return err
	}
	for _, a := range d.Data.Attachments {
		// get all cml attachments
		if a.Attachment.AttachmentType == "xml_transcription" {
			if a.Attachment.PresignedUrl != "" {
				urls <- a.Attachment.PresignedUrl
			}
		}
		// get all Facsimile attachments
		if a.Attachment.AttachmentType == "pdf" && a.DocumentAttachmentType == "Facsimile" {
			if a.Attachment.PresignedUrl != "" {
				urls <- a.Attachment.PresignedUrl
			}
		}
	}
	return nil
}

type sitemap struct {
	sitemapURL  string
	total       int
	currentPage int
	language    string
}

func newSitemap(currentPage int, language string) *sitemap {
	return &sitemap{
		sitemapURL:  "https://dodis.ch/search?q=*&c=Document&f=All&t=all&cb=doc",
		total:       currentPage + 2,
		currentPage: currentPage,
		language:    language,
	}
}

// queue next sitemap
func (s *sitemap) next(urls chan string) {
	// We handle state ourselves because dodis can return an invalid json. In this case
	// we had to restart the full process, because we loose the state.
	s.currentPage++
	if s.currentPage <= s.total {
		urls <- s.sitemapURL + "&p=" + strconv.Itoa(s.currentPage)
	}
}

// postSitemap handles all downloaded parts of the sitemap
// The func will extract all links to single docs and queue them
// the func will queue the next sitemap request
func (s *sitemap) postSitemap(urls chan string, m message) error {
	// log.Println("Process sitemap ")
	// Fetch next sitemap
	// This is a serial processing. Don't parallelize this. That will kill dodis
	s.next(urls)
	var sMap sitemapResponse
	err := json.Unmarshal(m.Content, &sMap)
	if err != nil {
		return err
	}
	// get all docs and queue them
	for _, d := range sMap.Data {
		if s.language == "all" {
			urls <- fmt.Sprintf("https://dodis.ch/%d", d.ID)
		} else {
			if d.LangCode == s.language {
				urls <- fmt.Sprintf("https://dodis.ch/%d", d.ID)
			}
		}
	}
	// all went good, we know the total pages
	s.total = sMap.TotalPages
	return nil
}

// sitemapResponse the overview api response from dodis
type sitemapResponse struct {
	Total int `json:"total"`
	Data  []struct {
		ID       int    `json:"id"`
		LangCode string `json:"langCode"`
	} `json:"data"`
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
}

// doc is the dodis API document response object
type doc struct {
	Data struct {
		RelatedTags []struct {
			TagId   int    `json:"tagId"`
			Name    string `json:"name"`
			MainTag bool   `json:"mainTag"`
		} `json:"relatedTags"`
		Attachments []struct {
			DocumentAttachmentType string `json:"documentAttachmentType"`
			Attachment             struct {
				AttachmentType string `json:"attachmentType"`
				PresignedUrl   string `json:"presignedUrl"`
			} `json:"attachment"`
		} `json:"attachments"`
	} `json:"data"`
}

// message is the command object we pass between our goroutines
type message struct {
	URL     string
	Path    string
	Type    documentType
	Content []byte
}

// documentType the type of document we want to process
type documentType int

// All documentTypes (enum)
const (
	FACSIMILE documentType = iota
	HTML
	JSON
	SITEMAP
)
