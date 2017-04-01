# gloomy
An ngram database


... a work in progress project


## Building an index

```
gloomy -ngram-size 3 create-index ./config.json
```

where *config.json* looks like this:

```json
{
    "verticalFilePath": "/path/to/a/vertical/file",
    "filterArgs": [],
    "ngramIgnoreStructs": [],
    "ngramStopStrings": [".", ":"],
    "ngramIgnoreStrings": ["\"", ","],
    "outDirectory": "/path/to/an/output/directory"
}
```

## Searching

... under development ...

## Additional functions

# Extracting sorted unique n-grams with frequencies

```
gloomy -ngram-size 3 extract-ngrams ./config.json
```
