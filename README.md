# InstaAppEngine

Take release of Go, AppEngine and some [gin](https://github.com/gin-gonic/gin). Drop a bit of [go-cache](https://github.com/patrickmn/go-cache). Shake well. Serve chilled. Enjoy.

## What is it?

AppEngine Go function for getting some metadata about user Instagram feed images.  This is quick adaptation of my [InstaAWS](https://github.com/chew-z/InstaAWS) AWS Lambda function.

## Status

Works well. Requires some further development.

## How to use it?

- set ENV variables (USERNAME, PASSWORD, GIN_MODE) in app.yaml
- deploy to AppEngine
```
  gcloud app deploy --version dev
```
- watch logs
```
  gcloud app logs tail -s default
```
- get JSON file with pictures metatdata from user feed
``` 
  curl -X GET -o instagram.json 'uri://appengine-ednpoint/instagram?limit=100' 
```

## TODO

- get metadata in chunks
- get random pictures

## Limitations



## What's interesting about it?

- I am using [goinsta](https://github.com/ahmdrz/goinsta) - good but it has it's limitations as Import/Export operations are hardcoded for storing session in file - not possible with Lambda. So my login() function is storing Instagram object in cache instead.
-