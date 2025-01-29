#!/bin/bash
# preprocessing
for filename in ./docs/*.html; do
  # extract text from html
  # echo "$filename" +"$(basename -- "$filename" .html).txt"
  ./dodo html2text "$filename" > "./docs/$(basename -- "$filename" .html).txt"
  # calc token frequency for each document
  ./dodo tf  "./docs/$(basename -- "$filename" .html).txt" >  "./docs/$(basename -- "$filename" .html)-tf.json"
done

