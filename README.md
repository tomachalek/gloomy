![Build status](https://travis-ci.org/tomachalek/gloomy.svg?branch=master)

# gloomy

An n-gram database written in Go, optimized for *write once read many* use.

... a work in progress project


## Building an index

Currently, *gloomy* expects a text source file to be compatible with [vertical format](https://www.sketchengine.co.uk/documentation/preparing-corpus-text/).

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

In the searching mode, a *gloomy.conf* file (by default in the working directory) is expected:

```json
{
    "dataPath": "/path/to/indices/data",
    "serverPort": 8090,
    "serverAddress": "127.0.0.1"
}
```

### command line mode

```
gloomy search corpname phrase
```

### HTTP server mode

Start a server:

```
gloomy search-service 
```

Test a client:

```
curl -XGET http://localhost:8090/search?corpus=susanne&q=but
```



## Additional functions

### Extracting sorted unique n-grams with frequencies

```
gloomy -ngram-size 3 extract-ngrams ./config.json
```
