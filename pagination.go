package jsonapidynamodb

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	PaginationCursorItemAttribute = "pagination_cursor" // references the page cursor in the page record.
	PaginationKeyItemAttribute    = "pagination_key"    // references the last evaluated key in the page record.
)

// Getter implements the DynamoDB client Get API.
type Getter interface {
	Get(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

// Putter implements the DynamoDB client Put API.
type Putter interface {
	Put(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// PaginationKeyProvider is responsible for generating DynamoDB API input structs
// that allows the client to retrieve and store page records to a dynamodb table.
type PaginationKeyProvider interface {
	// GenerateCursor generates a unique string that can be used by jsonapi clients
	// for pagination.
	GenerateCursor() string
	// GetPaginationKeyInput returns a DynamoDB GetItemInput struct that must contain
	// the following fields:
	//	- the table name
	//	- the record key associated with the provided cursor
	GetPaginationKeyInput(cursor string) dynamodb.GetItemInput
	// SetPaginationKeyInput returns a DynamoDB PutItemInput struct that must contain
	// the following fields:
	//	- the table name
	//  - the partial record item associated with the provided cursor
	SetPaginationKeyInput(cursor string) dynamodb.PutItemInput
}

// Record is a type alias for DynamoDB attribute value map.
type Record = map[string]types.AttributeValue

// Pagination structs provide a way to generate and store pagination cursors within
// a DynamoDB table.
type Pagination struct {
	Provider            PaginationKeyProvider // The provider to manage the pagination table.
	TimeToLiveAttribute string                // The optional attribute that stores the TTL attribute.
	Lifespan            time.Duration         // The optional lifespace of a page record. Default is 24 hours.
}

// NewPagination creates a new pagination instance.
func NewPagination(provider PaginationKeyProvider, options ...func(*Pagination)) Pagination {
	pagination := Pagination{
		Provider: provider,
		Lifespan: 24 * time.Hour,
	}

	for _, option := range options {
		option(&pagination)
	}

	return pagination
}

// GetLastEvaluatedKey retrieves the last evaluated key referenced by the provided page cursor.
func (p Pagination) GetLastEvaluatedKey(ctx context.Context,
	cursor string, getter Getter, options ...func(*dynamodb.Options)) (Record, error) {
	if cursor == "" {
		return nil, nil
	}

	input := p.Provider.GetPaginationKeyInput(cursor)

	output, err := getter.Get(ctx, &input, options...)
	if err != nil {
		return nil, fmt.Errorf("pagination: last evaluated key: %s", err)
	} else if output.Item == nil {
		return nil, nil
	}

	data, ok := output.Item[PaginationKeyItemAttribute].(*types.AttributeValueMemberB)
	if !ok {
		return nil, fmt.Errorf("pagination: page record is invalid")
	}

	lastEvaluatedKey := make(Record)
	buf := bytes.NewBuffer(data.Value)

	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&lastEvaluatedKey); err != nil {
		return nil, fmt.Errorf("pagination: decode page record: %s", err)
	}

	return lastEvaluatedKey, nil
}

// GetCursor generates a new page cursor from the last evaluated key that is returned from a
// DynamoDB query or scan. If the key is nil, then an empty string is returned.
func (p Pagination) GetCursor(ctx context.Context,
	lastEvaluatedKey Record, putter Putter, options ...func(*dynamodb.Options)) (string, error) {
	if lastEvaluatedKey == nil {
		return "", nil
	}

	cursor := p.Provider.GenerateCursor()
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(lastEvaluatedKey); err != nil {
		return "", fmt.Errorf("pagination: cursor encode page key: %s", err)
	}

	input := p.Provider.SetPaginationKeyInput(cursor)
	input.Item[PaginationKeyItemAttribute] = &types.AttributeValueMemberB{Value: buf.Bytes()}
	input.Item[PaginationCursorItemAttribute] = &types.AttributeValueMemberS{Value: cursor}

	if p.TimeToLiveAttribute != "" {
		expires := time.Now().UTC().Add(p.Lifespan)
		value := strconv.FormatInt(expires.Unix(), 10)
		input.Item[p.TimeToLiveAttribute] = &types.AttributeValueMemberN{Value: value}
	}

	_, err := putter.Put(ctx, &input, options...)
	if err != nil {
		return "", fmt.Errorf("pagination: cursor put: %s", err)
	}

	return cursor, nil
}
