package stores

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const batchSize = 25 // DynamoDB allows a maximum of 25 items per batch write
const pkURLPrefix = "url#"
const primaryKey = "pk"
const secondaryKey = "sk"
const skURLPrefix = "url#"
const tableName = "table-1"

type Table1 struct {
	client    *dynamodb.Client
	tableName string
	timeout   time.Duration
}

type table1Record struct {
	table1RecordPrimaryKey
}

type table1RecordPrimaryKey struct {
	PartitionKey string `dynamodbav:"pk"`
	SortKey      string `dynamodbav:"sk"`
}

func newURLRecord(url string) table1Record {
	recPk := newURLRecordPrimaryKey(url)
	return table1Record{
		recPk,
	}
}

func newURLRecordPrimaryKey(url string) table1RecordPrimaryKey {
	return table1RecordPrimaryKey{
		PartitionKey: fmt.Sprintf("%s%s", pkURLPrefix, url),
		SortKey:      fmt.Sprintf("%s%s", skURLPrefix, url),
	}
}

func InitializeTable1(ctx context.Context, tableArn string, timeout time.Duration, endpointURL *string) (*Table1, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading default config: %v", err)
	}
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpointURL != nil {
			o.BaseEndpoint = aws.String(*endpointURL)
		}
	})
	tableName, err := getTableName(ctx, client, tableArn, timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting table-name for table-arn='%s': %v", tableArn, err)
	}
	return &Table1{client: client, tableName: tableName, timeout: timeout}, nil
}

func getTableName(ctx context.Context, client *dynamodb.Client, tableArn string, timeout time.Duration) (string, error) {
	tablePaginator := dynamodb.NewListTablesPaginator(client, &dynamodb.ListTablesInput{})
	for tablePaginator.HasMorePages() {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		tablesOutput, err := tablePaginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("error listing tables: %v", err)
		}
		for _, tableName := range tablesOutput.TableNames {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			describeOutput, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
			if err != nil {
				return "", fmt.Errorf("error describing table with table-name='%s': %v", tableName, err)
			}
			_tableArn := strings.TrimSpace(*describeOutput.Table.TableArn)
			if _tableArn == tableArn {
				return tableName, nil
			}
		}
	}
	return "", fmt.Errorf("could not find table-name for table-arn='%s'", tableArn)
}

func (table1 *Table1) PutURL(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, table1.timeout)
	defer cancel()
	record := newURLRecord(url)
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("error creating item for record: %v", err)
	}
	_, err = table1.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &table1.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("error on PutItem into ddb table: %v", err)
	}
	return nil
}

func (table1 *Table1) HasURL(ctx context.Context, url string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, table1.timeout)
	defer cancel()
	urlRecordPrimaryKey := newURLRecordPrimaryKey(url)
	keyInput, err := attributevalue.MarshalMap(urlRecordPrimaryKey)
	if err != nil {
		return false, fmt.Errorf("error on MarshalMap over urlRecordPrimaryKey: %v", err)
	}
	item, err := table1.client.GetItem(ctx, &dynamodb.GetItemInput{
		Key:       keyInput,
		TableName: aws.String(table1.tableName),
	})
	if err != nil {
		return false, fmt.Errorf("error on GetItem: %v", err)
	}
	hasUrl := len(item.Item) > 0
	return hasUrl, nil
}

func (table1 *Table1) DeleteAll(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, table1.timeout)
	defer cancel()
	scanPaginator := dynamodb.NewScanPaginator(table1.client, &dynamodb.ScanInput{
		TableName: &table1.tableName,
	})

	var writeRequests []types.WriteRequest = make([]types.WriteRequest, 0, batchSize)

	for scanPaginator.HasMorePages() {
		ctx, cancel := context.WithTimeout(ctx, table1.timeout)
		defer cancel()
		scanPage, err := scanPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("error fetching next scan page: %v", err)
		}

		for _, item := range scanPage.Items {
			writeRequests = append(writeRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"pk": item["pk"],
						"sk": item["sk"],
					},
				},
			})

			// Batch write when we reach batchSize
			if len(writeRequests) == batchSize {
				if err := table1.batchWrite(ctx, writeRequests); err != nil {
					return fmt.Errorf("error during batch write: %v", err)
				}
				writeRequests = make([]types.WriteRequest, 0, batchSize)
			}
		}
	}

	// Write remaining items if any
	if len(writeRequests) > 0 {
		if err := table1.batchWrite(ctx, writeRequests); err != nil {
			return fmt.Errorf("error during final batch write: %v", err)
		}
	}

	return nil
}

func (table1 *Table1) batchWrite(ctx context.Context, writeRequests []types.WriteRequest) error {
	_, err := table1.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			table1.tableName: writeRequests,
		},
	})
	return err
}
