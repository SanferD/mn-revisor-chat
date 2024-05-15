#! /bin/bash

QUEUE_NAME=url-to-crawl
TABLE_NAME=table-1

# confirm that localstack exists
if ! command -v localstack &> /dev/null
then
    echo "LocalStack is not installed. Please install LocalStack to use its services. Visit https://localstack.cloud/ for installation instructions."
    exit 1
fi

# confirm that awslocal exists
if ! command -v awslocal &> /dev/null
then
    echo "awslocal command not found. Please ensure LocalStack is properly installed. Refer to https://localstack.cloud/ for more details."
    exit 1
fi

# Test the Docker connection by retrieving the version
docker version > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "Failed to connect to Docker."
  exit 1
fi

# start localstack if it isn't already running
MAIN_CONTAINER_NAME="localstack-pro" localstack start -d || echo "localstack already running"
MAIN_CONTAINER_NAME="localstack-pro" localstack wait

# Check if the table exists and get the ARN, create the table if it does not exist
TABLE_1_ARN=$(awslocal dynamodb describe-table --table-name $TABLE_NAME | jq -r '.Table.TableArn' 2>/dev/null)
if [ -z "$TABLE_1_ARN" ]; then
    echo "Table does not exist. Creating table..."
    TABLE_1_ARN=$(awslocal dynamodb create-table --table-name $TABLE_NAME \
        --attribute-definitions AttributeName=pk,AttributeType=S AttributeName=sk,AttributeType=S \
        --key-schema AttributeName=pk,KeyType=HASH AttributeName=sk,KeyType=RANGE \
        --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
        | jq -r '.TableDescription.TableArn')
    echo "Table created. ARN: $TABLE_1_ARN"
else
    echo "Table already exists. ARN: $TABLE_1_ARN"
fi

# Check if the queue exists and get the URL, otherwise create the queue
QUEUE_URL=$(awslocal sqs get-queue-url --queue-name $QUEUE_NAME | jq -r '.QueueUrl' 2>/dev/null)

if [ -z "$QUEUE_URL" ]; then
    echo "Queue does not exist. Creating queue..."
    QUEUE_URL=$(awslocal sqs create-queue --queue-name $QUEUE_NAME | jq -r '.QueueUrl')
    echo "Queue created. URL: $QUEUE_URL"
else
    echo "Queue already exists."
fi

# Get the ARN of the queue using the queue URL
QUEUE_ARN=$(awslocal sqs get-queue-attributes --queue-url $QUEUE_URL --attribute-names QueueArn | jq -r '.Attributes.QueueArn')

if [ -z "$QUEUE_ARN" ]; then
    echo "Failed to get queue ARN."
else
    echo "Queue ARN: $QUEUE_ARN"
fi

# create s3 bucket if it doesn't already exist
BUCKET_NAME="mn-revisor-chat-dev"
if awslocal s3api head-bucket --bucket $BUCKET_NAME 2>/dev/null; then
    echo "Bucket $BUCKET_NAME already exists."
else
    echo "Bucket $BUCKET_NAME does not exist. Creating bucket..."
    awslocal s3 mb "s3://${BUCKET_NAME}/"
    echo "Bucket created."
fi

# echo the settings
echo
echo
echo "BUCKET_NAME=\"$BUCKET_NAME\""
echo "LOCAL_ENDPOINT=\"http://localhost:4566/\""
echo "RAW_PATH_PREFIX=\"raw\""
echo "TABLE_1_ARN=\"$TABLE_1_ARN\""
echo "URL_SQS_ARN=\"$QUEUE_ARN\""
