package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	cmd := os.Args[1:2][0]
	if cmd == "scrap" {
		// define default params
		currentPage := 1
		language := "all"
		// check for page
		if len(os.Args) >= 4 {
			var err error
			currentPage, err = strconv.Atoi(os.Args[2:3][0])
			if err != nil {
				log.Fatal("Second argument have to be a number")
			}
		}
		// check for language
		if len(os.Args) >= 5 {
			language = os.Args[3:4][0]
		}
		log.Println("Scrap dodis document for language=" + language + " from the api starting with sitemap page=" + strconv.Itoa(currentPage))
		scrapDocs(currentPage, language)
	} else if cmd == "clean" {
		clean()
	} else if cmd == "tf" {
		fmt.Println(newTokenFrequency().FromFile(os.Args[2:3][0]))
	} else if cmd == "xml2Text" {
		fmt.Println(xmlTotext(os.Args[2:3][0]))
	} else if cmd == "html2text" {
		fmt.Println(htmlToText(os.Args[2:3][0]))
	} else if cmd == "2spacy" {
		paths := os.Args[2:4]
		text := ""
		if strings.Contains(paths[0], ".xml") {
			text = xmlTotext(paths[0])
		} else if strings.Contains(paths[0], ".html") {
			text = htmlToText(paths[0])
		} else {
			log.Fatal("Can't find the xml or html file")
		}
		t, err := tags(paths[1])
		if err != nil {
			log.Fatalf("Can't read tags, %v", err)
		}
		fmt.Println(string(toSpacy(text, t)))
	} else {
		log.Println("Help")
		log.Println("usage: dodo cmd [args] [de|fr|it|all]")
		log.Println("cmd scarp [pagenumber]: scrap the sitemap, all the german documents, pdfs and transcription from the dodis api. ")
		log.Println("cmd clean: remove all documents without transcription from docs folder")
		log.Println("cmd xmlTotext doc.xml: transform a dodis xml file into txt file")
		log.Println("cmd html2text doc.html: transform a dodis html file into txt file")
		log.Println("cmd 2spacy doc.(xml|html) doc.json: combine a dodis html or xml document with the its tags and return a spacy ready json")
	}
}
