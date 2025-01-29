package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/jaytaylor/html2text"
	"github.com/microcosm-cc/bluemonday"
)

// xmlTotext extracts the body from a TEI XML file and return it cleaned from all xml tags
func xmlTotext(file string) string {
	text, err := os.ReadFile(file)
	if err != nil {
		log.Println("Can't find xml file")
		return ""
	}
	d, err := xmlquery.Parse(bytes.NewReader(text))
	if err != nil {
		log.Println(err)
		return ""
	}
	list, err := xmlquery.Query(d, "//body")
	if err != nil {
		log.Println(err)
		return ""
	}
	if list == nil {
		return ""
	}
	xml := list.OutputXML(true)
	// remove special TEI tags, that would make our text dump
	re := regexp.MustCompile(`<note>[\s\S]*?<\/note>`)
	xml = re.ReplaceAllString(xml, "")
	re = regexp.MustCompile(`<head>[\s\S]*?<\/head>`)
	xml = re.ReplaceAllString(xml, "")
	// remove all TEI-tags
	p := bluemonday.StrictPolicy()
	p.AddSpaceWhenStrippingTag(true)
	return p.Sanitize(xml)
}

// toSpacy turns a text and some tags into a spacy json
func toSpacy(text string, tags []string) []byte {
	spacy := spacyInput{
		Text: text,
		Tags: tags,
	}
	data, _ := json.Marshal(spacy)
	return data
}

// htmlToText turn a htl into a text file
func htmlToText(file string) string {
	text, err := os.ReadFile(file)
	if err != nil {
		log.Println("Can't find html file")
		return ""
	}
	// <a class="note"
	re := regexp.MustCompile(`<a class="note".*?>.*?<\/a>`)
	reText := re.ReplaceAllString(string(text), "")
	re = regexp.MustCompile(`<div class="footnotes">[\s\S]*?<\/div>`)
	reText = re.ReplaceAllString(reText, "")
	rawText, _ := html2text.FromString(reText, html2text.Options{PrettyTables: false, OmitLinks: true, TextOnly: true})
	return rawText
}

// clean removes unused json docs. We just keep those with a transcription
func clean() {
	files, err := os.ReadDir("./docs/")
	if err != nil {
		log.Fatal(err)
	}

	keep := []string{}
	for _, file := range files {
		info, _ := file.Info()
		if strings.Contains(info.Name(), ".html") {
			keep = append(keep, strings.Replace(info.Name(), ".html", "", 1))
		}
	}
	fmt.Println("Keep files " + strconv.Itoa(len(keep)))
	// not very efficient
	for _, file := range files {
		info, _ := file.Info()
		keepIt := false
		for _, k := range keep {
			pattern := fmt.Sprintf(`^%s(-fq\.json|\.html|.xml|\.json)$`, k)
			if ok, _ := regexp.MatchString(pattern, info.Name()); ok {
				keepIt = true
				break
			}
		}
		if !keepIt {
			fmt.Println("Delete file " + info.Name())
			err = os.Remove(path.Join("docs", info.Name()))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// spacyInput is the struct we construct to prepare data for spacy
type spacyInput struct {
	Text string   `json:"text"`
	Tags []string `json:"tags"`
}
