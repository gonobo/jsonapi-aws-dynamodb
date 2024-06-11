package jsonapidynamodb

import (
	"errors"
	"fmt"

	"github.com/gonobo/jsonapi/query"
	"github.com/gonobo/jsonapi/query/filter"
)

var (
	ErrDynamoDB = errors.New("dynamodb integration error")
)

func DynamoDBIntegrationError(format string, args ...any) error {
	cause := fmt.Sprintf(format, args...)
	return fmt.Errorf("%w: %s", ErrDynamoDB, cause)
}

type IntersectsListCriteria struct {
	List []string
}

func (i IntersectsListCriteria) String() string {
	return fmt.Sprintf("INTERSECTS_LIST(%s)", i.List)
}

func (i *IntersectsListCriteria) ApplyFilterEvaluator(e query.FilterEvaluator) error {
	return e.EvaluateCustomFilter(i)
}

func (i IntersectsListCriteria) Transformer() filter.TransformerFunc {
	return func(*query.Filter) (query.FilterExpression, error) {
		return &i, nil
	}
}
