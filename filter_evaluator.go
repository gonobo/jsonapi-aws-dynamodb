package jsonapidynamodb

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/gonobo/jsonapi/query"
)

type FilterEvaluator struct {
	HashKey string
	Builder expression.ConditionBuilder
}

func (d *FilterEvaluator) EvaluateAndFilter(e *query.AndFilter) error {
	var left, right FilterEvaluator
	left.HashKey = d.HashKey
	right.HashKey = d.HashKey

	lefterr := query.EvaluateFilter(&left, e.Left)
	righterr := query.EvaluateFilter(&right, e.Right)

	d.Builder = expression.And(left.Builder, right.Builder)
	return errors.Join(lefterr, righterr)
}

func (d *FilterEvaluator) EvaluateOrFilter(e *query.OrFilter) error {
	var left, right FilterEvaluator
	left.HashKey = d.HashKey
	right.HashKey = d.HashKey

	lefterr := query.EvaluateFilter(&left, e.Left)
	righterr := query.EvaluateFilter(&right, e.Right)

	d.Builder = expression.Or(left.Builder, right.Builder)
	return errors.Join(lefterr, righterr)
}

func (d *FilterEvaluator) EvaluateNotFilter(e *query.NotFilter) error {
	var value FilterEvaluator
	value.HashKey = d.HashKey

	err := query.EvaluateFilter(&value, e.Value)

	d.Builder = expression.Not(value.Builder)
	return err
}

func (d *FilterEvaluator) EvaluateIdentityFilter() error {
	d.Builder = expression.AttributeExists(expression.Name(d.HashKey))
	return nil
}

func (d *FilterEvaluator) EvaluateFilter(criteria *query.Filter) error {
	switch criteria.Condition {
	case query.Equal:
		d.Builder = expression.Equal(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
		return nil
	case query.NotEqual:
		d.Builder = expression.NotEqual(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
		return nil
	case query.Contains:
		d.Builder = expression.Contains(
			expression.Name(criteria.Name),
			criteria.Value,
		)
	case query.LessThan:
		d.Builder = expression.LessThan(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.LessThanEqual:
		d.Builder = expression.LessThanEqual(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.GreaterThan:
		d.Builder = expression.GreaterThan(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.GreaterThanEqual:
		d.Builder = expression.GreaterThanEqual(
			expression.Name(criteria.Name),
			expression.Value(criteria.Value),
		)
	case query.StartsWith:
		d.Builder = expression.BeginsWith(
			expression.Name(criteria.Name),
			criteria.Value,
		)
	}

	return fmt.Errorf("unknown or invalid token: %s", criteria)
}

type DynamoDBCustomCondition interface {
	Condition() expression.ConditionBuilder
}

func (d *FilterEvaluator) EvaluateCustomFilter(value any) error {
	if custom, ok := value.(DynamoDBCustomCondition); ok {
		d.Builder = custom.Condition()
		return nil
	}
	return DynamoDBIntegrationError("custom condition not supported")
}
