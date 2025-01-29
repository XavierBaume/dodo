package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/dchest/stemmer/german"
	"golang.org/x/exp/slices"
)

// tokenFrequency is the struct to calc tf
type tokenFrequency struct {
	stopWords          []string
	stemming           bool
	minCharacterLength int
	toLower            bool
	relative           bool
	binary             bool
}

// newTokenFrequency reads the stop words and returns a struct to processdata
func newTokenFrequency() *tokenFrequency {
	stopword, err := os.ReadFile("stopwords.txt")
	if err != nil {
		log.Fatalln("Can't load stopwords.txt file")
	}
	return &tokenFrequency{
		stopWords:          strings.Fields(string(stopword)),
		stemming:           true,
		toLower:            true,
		minCharacterLength: 4,
		relative:           false,
		binary:             true,
	}
}

// split takes a []byte input and return a slice of tokens
func spilt(text []byte) [][]byte {
	return bytes.Fields(text)
}

// stem a word to its root
func stem(word string) string {
	stemmer := german.Stemmer
	return stemmer.Stem(word)
}

// Parse a string into a json of token frequency
func (t *tokenFrequency) Parse(text string) string {
	// remove all punctuation
	re := regexp.MustCompile(`[\.:;,\?!\+\/]`)
	refinedText := re.ReplaceAllString(text, " ")
	wf := t.wordFrequency(spilt([]byte(refinedText)))
	data, _ := json.Marshal(wf)
	return string(data)
}

// FromFile reads a file and calculate the token frequency
func (t *tokenFrequency) FromFile(file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalln("Cant read file to calc token frequency")
	}
	return t.Parse(string(data))
}

// tokenFrequency takes a word slice and count its frequency. Before counting, it transforms the word (toLower,stemming, exclude numbers, exclude too short words)
func (t *tokenFrequency) wordFrequency(words [][]byte) map[string]float64 {
	sort.Slice(words, func(i, j int) bool {
		return bytes.Compare(words[i], words[j]) == -1
	})
	localWordList := map[string]float64{}

	// count each word
	for _, word := range words {
		w := string(word)
		if t.toLower {
			w = strings.ToLower(w)
		}
		// exclude stopwords before stemming
		if slices.Contains(t.stopWords, w) {
			continue
		}
		if t.stemming {
			w = stem(w)
			// exclude stopwords after stemming
			if slices.Contains(t.stopWords, w) {
				continue
			}
		}
		// exclude too sorts words
		if len(w) < t.minCharacterLength {
			continue
		}
		// exclude numbers only
		re := regexp.MustCompile(`[0-9\.+]`)
		if re.MatchString(w) {
			continue
		}
		// count frequency
		if _, ok := localWordList[w]; ok {
			localWordList[w]++
		} else {
			localWordList[w] = 1
		}
	}
	if t.relative {
		for key, value := range localWordList {
			localWordList[key] = value / float64(len(localWordList))
		}
	}
	if t.binary {
		for key, _ := range localWordList {
			localWordList[key] = 1
		}

	}
	return localWordList
}
