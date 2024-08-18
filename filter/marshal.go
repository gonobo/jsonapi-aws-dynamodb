package filter

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/gonobo/jsonapi/v2/query"
)

type Marshaler struct {
	HashKey string
	builder expression.ConditionBuilder
}

func NewFilterMarshaler(hashKey string) *Marshaler {
	return &Marshaler{HashKey: hashKey}
}

func (m *Marshaler) EvaluateAndFilter(e *query.AndFilter) error {
	var left, right Marshaler
	left.HashKey = m.HashKey
	right.HashKey = m.HashKey

	lefterr := query.EvaluateFilter(&left, e.Left)
	righterr := query.EvaluateFilter(&right, e.Right)

	m.builder = expression.And(left.builder, right.builder)
	return errors.Join(lefterr, righterr)
}

func (m *Marshaler) EvaluateOrFilter(e *query.OrFilter) error {
	var left, right Marshaler
	left.HashKey = m.HashKey
	right.HashKey = m.HashKey

	lefterr := query.EvaluateFilter(&left, e.Left)
	righterr := query.EvaluateFilter(&right, e.Right)

	m.builder = expression.Or(left.builder, right.builder)
	return errors.Join(lefterr, righterr)
}

func (m *Marshaler) EvaluateNotFilter(e *query.NotFilter) error {
	var value Marshaler
	value.HashKey = m.HashKey

	err := query.EvaluateFilter(&value, e.Expression)

	m.builder = expression.Not(value.builder)
	return err
}

func (m *Marshaler) EvaluateIdentityFilter() error {
	m.builder = expression.AttributeExists(expression.Name(m.HashKey))
	return nil
}

func (m *Marshaler) EvaluateFilter(criteria *query.Filter) error {
	condition := query.FilterCondition(criteria.Condition)
	switch condition {
	case query.Equal:
		m.builder = expression.Equal(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
		return nil
	case query.NotEqual:
		m.builder = expression.NotEqual(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
		return nil
	case query.Contains:
		m.builder = expression.Contains(
			expression.Name(criteria.Name),
			criteria.Value,
		)
	case query.LessThan:
		m.builder = expression.LessThan(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.LessThanEqual:
		m.builder = expression.LessThanEqual(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.GreaterThan:
		m.builder = expression.GreaterThan(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.GreaterThanEqual:
		m.builder = expression.GreaterThanEqual(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.StartsWith:
		m.builder = expression.BeginsWith(
			expression.Name(criteria.Name),
			criteria.Value,
		)
	}

	return fmt.Errorf("unknown or invalid token: %s", criteria)
}

type CustomFilter interface {
	Condition() (expression.ConditionBuilder, error)
}

func (m *Marshaler) EvaluateCustomFilter(value any) error {
	if custom, ok := value.(CustomFilter); ok {
		condition, err := custom.Condition()
		m.builder = condition
		return err
	}
	return fmt.Errorf("custom condition not supported: %v", value)
}
