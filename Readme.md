# DoDo

This is a small dodis helper library. It's a collection of scripts to scrap the dodis api, to clean dodis data and to preprocess dodis document for machine learning.

## Usage

```bash
# fetch all dodis HTML, pdf and documents form the server 
# this creates a folder ./docs in which yo can find all the documents
dodo scrap

# turn a dodis html (TEI) into text
dodo html2text ./docs/3.html

# turn a dodis XML (TEI) into text
dodo xml2text ./docs/3.xml

# turn a dodis html and a dodis json document into a spacy json
dodo 2spacy ./docs/3.html ./docs/3.json

# calculate the token frequency of a text document
dodo tf ./docs/3.txt

# remove all unused data from docs (all the documents without a transcription)
dodo clean
```

For staple processing you can write a small bash script. ave a look into [preprocess.sh](./preprocess.sh) 

```bash
for filename in ./docs/*.html; do
  dodo html2text $filename "$(basename -- "filename" .html).txt"
done
```
## scrap

```bash
dodo scrap [currentPage] [langauges (de|fr|it|all)]
```

Scrap will crawl the Dodis api and saves all documents, pdfs and html transcriptions. This process will spin up a crawler and fetch documents concurrently with some PolitePolicy (sleep 400ms) between each request.

**Timeout:**

Dodis may return a timeout if the load on the server is too high and end the scrapping process. You can resume the process with the `currentPage` argument. 

**Language**

By default, dodo will scrap all languages. If you need just one, you can use the second argument to narrow down the scraping.

```bash
dodo scrap [currentPage] [language]
```

**Args**
* currentPage Startpoint for the scrapping Default=0
* language The language of the documents Default=all

**Features**
* Fetch the dodis "sitemap" and saves it
* Extract all the links from the sitemap and fetch the json docs, the html transcription and the pdf facsimile

**Problems**
* Most documents has no html-transcription. So we have a small base
* The Dodis-API makes it not so easy to scrap all those things. We need signed requests to get PDF and HTML. From time to time it fails. But it's doable!

_Hint: this will download mire than 20Gb of data_
## html2text

Turn a dodis html transcription into a simple text format. We can't use linux tool [html2text](https://salsa.debian.org/debian/html2text), because we have to do some custom transformation. We wan't to remove all the footnotes from the html first. They mess up the outcome.

```bash
dodo html2text ./doccs/3.html
```

**Features**
* Remove footnotes from the html
* Turn html into text
* Turn paragraph's into line breaks
* Remove everything else

## xml2text

Turn a dodis xml transcription (TEI) into a simple text format. We can't use linux onboard tool, because we have to do some custom transformation. We wan't to remove all the notes and headings from the xml first. They mess up the outcome.

```bash
dodo xml2text ./docs/3.xml
```

**Features**
* Remove footnotes, notes and headings from the html
* Turn TEI-XML into text
* Remove everything else

**Todo**
* Preserve linebreaks

## tf
Simply calculate the token frequency from a text document and return it as a json. This is a simple task and can be done with other tools (f.e. panda). But it helped to have it right here. We do stemming (german), use `stopwords.txt` and do some tricks for dodis documents.

```bash
dodo tf ./docs/3.txt
```
**Features**
* Calculate token-frequency for a text docuemnt
* Stem german words
* Use stopwords to exclude unspecific words
* Remove numbers
* remove tokens smaller han 4 characters

## 2spacy

To load data into spacy we have to preprocess them. **This will not produce spacy binary format**, just some json we can load in python. Perhaps we ove this to the python world.

```bash
dodo 2spacy ./docs/3.html ./docs/3.json 
```
**Features**
* Generate a json with a `text` and a `tags` key. In `text` you can find the content of the document and in `tags` the corresponding MainTags from dodis. 

## clean

The scrapper loads a ton of data. Just for some of this data we have a transcription. Most further processing just deals with the transcription. The clean command deletes everything else from `./docs`

```bash
dodo clean
```



