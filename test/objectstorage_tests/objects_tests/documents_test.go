package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/codenotary/immudb/test/objectstorage_tests/immudbhttpclient/immudbdocuments"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateDocument(t *testing.T) {
	client := getAuthorizedDocumentsClient()
	id := uuid.New()
	documentId := id.String()
	documentToInsert := make(map[string]interface{})
	documentToInsert["_id"] = id.String()
	documentToInsert["name"] = "John"
	documentToInsert["surname"] = "Doe"
	documentToInsert["age"] = 30
	documentsToInsert := []map[string]interface{}{documentToInsert}
	collectionName := CreateAndGetStandardTestCollection(client)

	req := immudbdocuments.DocumentschemaDocumentInsertRequest{
		Collection: &collectionName,
		Document:   &documentsToInsert,
	}
	response, _ := client.DocumentServiceDocumentInsertWithResponse(context.Background(), req)
	assert.True(t, response.StatusCode() == 200)
	page := int64(1)
	perPage := int64(100)
	operator := immudbdocuments.EQ
	fieldName := "_id"
	query := []immudbdocuments.DocumentschemaDocumentQuery{
		{
			Field:    &fieldName,
			Value:    &documentId,
			Operator: &operator,
		},
	}
	searchReq := immudbdocuments.DocumentschemaDocumentSearchRequest{
		Collection: &collectionName,
		Page:       &page,
		PerPage:    &perPage,
		Query:      &query,
	}
	searchResponse, _ := client.DocumentServiceDocumentSearchWithResponse(context.Background(), searchReq)
	fmt.Println(searchResponse.StatusCode())
	assert.True(t, searchResponse.StatusCode() == 200)
	documents := *searchResponse.JSON200.Results
	first := documents[0]
	assert.True(t, first["_id"] == documentId)
	assert.True(t, first["age"] == 30)
	assert.True(t, first["name"] == "John")
	assert.True(t, first["surname"] == "Doe")

}