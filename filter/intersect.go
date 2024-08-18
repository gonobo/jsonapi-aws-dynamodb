package filter

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/gonobo/jsonapi/v1/query"
	"github.com/gonobo/jsonapi/v1/query/filter"
)

type IntersectsListFilter struct {
	Name string
	List []string
}

func (i IntersectsListFilter) String() string {
	return fmt.Sprintf("INTERSECTS_LIST(%s, %s)", i.Name, i.List)
}

func (i *IntersectsListFilter) ApplyFilterEvaluator(e query.FilterEvaluator) error {
	return e.EvaluateCustomFilter(i)
}

func (i IntersectsListFilter) Transformer() filter.TransformerFunc {
	return func(*query.Filter) (query.FilterExpression, error) {
		return &i, nil
	}
}

func (i IntersectsListFilter) Condition() (expression.ConditionBuilder, error) {
	if len(i.List) == 0 {
		return expression.ConditionBuilder{}, fmt.Errorf("intersect list is empty")
	}

	var condition expression.ConditionBuilder
	name := expression.Name(i.Name)
	for _, v := range i.List {
		if !condition.IsSet() {
			condition = expression.Contains(name, v)
		}
		condition = condition.Or(expression.Contains(name, v))
	}
	return condition, nil
}
