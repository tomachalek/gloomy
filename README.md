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
    "outDirectory": "/path/to/an/output/directory",
    "args": {
      "doc.file": "col8",
      "doc.n": "col8",
      "head.type": "col8"
  }
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
curl -XGET http://localhost:8090/search?corpus=susanne&q=from
```


### Query syntax

The current version supports only a search by the first token.

Exact search:

```
gloomy search susanne absolute
```

... searches for all the n-grams with the first token equal to *absolute*.


Search by a prefix:

```
gloomy search susanne abs*
```

... searches for all the n-grams where the first token starts with *abs\**


### Metadata retrieval

**Command line**:

```
gloomy seaerch --attrs doc.file,doc.n susanne absolute
```

In **HTTP server mode** use multi-value attribute:

```
http://localhost:8090/search?corpus=susanne&q=from&attrs=doc.file&attrs=doc.n
```


## Additional functions

### Extracting sorted unique n-grams with frequencies

```
gloomy -ngram-size 3 extract-ngrams ./config.json
```
