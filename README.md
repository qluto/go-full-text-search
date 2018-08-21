It's just exercise.

# Boot server
```
go run main.go
```

# Interface
## Put (Update) a document
```
curl -X PUT -d body="Lorem ipsum dolor sit amet Lorem" localhost:8080/document/{anyDocId}
```

## Delete a document
```
curl -X DELETE localhost:8080/document/{anyDocId}
```

## Search documents
```
curl -X GET localhost:8080/search?q="foo+AND+bar"
```
