#!/bin/bash
for filename in ./docs/*tf.json; do
  # extract text from html
  cp "./docs/$(basename -- "$filename" -tf.json).json" ~/dds/python-tagger/docs/
done
