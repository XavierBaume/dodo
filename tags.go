package main

import (
	"encoding/json"
	"os"
)

// tags reads all the tags for a specific document
func tags(filename string) ([]string, error) {
	finalTags := []string{}
	file, err := os.ReadFile(filename)
	if err != nil {
		return finalTags, err
	}
	var d doc
	err = json.Unmarshal(file, &d)
	if err != nil {
		return finalTags, err
	}
	for _, t := range d.Data.RelatedTags {
		if t.MainTag {
			finalTags = append(finalTags, t.Name)
		}
	}
	return finalTags, nil
}
