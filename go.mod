module github.com/gonobo/jsonapi/v1/extra/dynamodb

go 1.22.3

require github.com/aws/aws-sdk-go-v2/service/dynamodb v1.31.0 // indirect

require (
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.13.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.20.3 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.7.11
	github.com/aws/smithy-go v1.20.1 // indirect
)
