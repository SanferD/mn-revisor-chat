#! /bin/bash

EMBEDDING_MODEL_ID="amazon.titan-embed-text-v2:0"
OPENSEARCH_DOMAIN="https://localhost:9200"
OPENSEARCH_INDEX_NAME="subdivisions-knn"
OPENSEARCH_PASSWORD="admin"
OPENSEARCH_USERNAME="admin"
SINCH_API_TOKEN="ba9bb80d5ebb490f8f08056ddf754cb5"
SINCH_SERVICE_ID="5a65ba33a6014f17a124c98bd8ee7b28"
SINCH_VIRTUAL_PHONE_NUMBER="+13203476019"

# Check if the --overwrite-settings flag is set
OVERWRITE_SETTINGS=false
for arg in "$@"; do
    if [ "$arg" == "--overwrite-settings" ]; then
        OVERWRITE_SETTINGS=true
    fi
done

# Function to write settings to settings.env
write_settings() {
    echo "Writing settings to settings.env..."
    cat <<EOL > ./settings.env
MAIN_BUCKET_NAME="$MAIN_BUCKET_NAME"
CHUNK_PATH_PREFIX="chunk"
LOCAL_ENDPOINT="http://localhost:4566/"
RAW_EVENTS_SQS_ARN="$RAW_EVENTS_SQS_ARN"
RAW_PATH_PREFIX="raw"
TABLE_1_ARN="$TABLE_1_ARN"
URL_SQS_ARN="$URL_SQS_ARN"
EMBEDDING_MODEL_ID="$EMBEDDING_MODEL_ID"
OPENSEARCH_DOMAIN="$OPENSEARCH_DOMAIN"
OPENSEARCH_INDEX_NAME="$OPENSEARCH_INDEX_NAME"
OPENSEARCH_PASSWORD="$OPENSEARCH_PASSWORD"
OPENSEARCH_USERNAME="$OPENSEARCH_USERNAME"
SINCH_API_TOKEN="$SINCH_API_TOKEN"
SINCH_SERVICE_ID="$SINCH_SERVICE_ID"
SINCH_VIRTUAL_PHONE_NUMBER="$SINCH_VIRTUAL_PHONE_NUMBER"
EOL
    echo "Settings written to ./settings.env"
}

get_or_create_queue_arn() {
    local QUEUE_NAME=$1
    local QUEUE_URL
    local URL_SQS_ARN

    # Check if the queue exists and get the URL, otherwise create the queue
    QUEUE_URL=$(awslocal sqs get-queue-url --queue-name $QUEUE_NAME | jq -r '.QueueUrl' 2>/dev/null)

    if [ -z "$QUEUE_URL" ]; then
        echo "Queue does not exist. Creating queue..." >&2
        QUEUE_URL=$(awslocal sqs create-queue --queue-name $QUEUE_NAME | jq -r '.QueueUrl')
        echo "Queue created. URL: $QUEUE_URL" >&2
    else
        echo "Queue already exists." >&2
    fi

    # Get the ARN of the queue using the queue URL
    URL_SQS_ARN=$(awslocal sqs get-queue-attributes --queue-url $QUEUE_URL --attribute-names QueueArn | jq -r '.Attributes.QueueArn')

    if [ -z "$URL_SQS_ARN" ]; then
        echo "Failed to get queue ARN." >&2
        return 1
    else
        echo "Queue ARN: $URL_SQS_ARN" >&2
        echo $URL_SQS_ARN
    fi
}

get_or_create_table_arn() {
    local TABLE_NAME=$1
    local TABLE_ARN

    # Check if the table exists and get the ARN, otherwise create the table
    TABLE_ARN=$(awslocal dynamodb describe-table --table-name $TABLE_NAME | jq -r '.Table.TableArn' 2>/dev/null)

    if [ -z "$TABLE_ARN" ]; then
        echo "Table does not exist. Creating table..." >&2
        TABLE_ARN=$(awslocal dynamodb create-table --table-name $TABLE_NAME \
            --attribute-definitions AttributeName=pk,AttributeType=S AttributeName=sk,AttributeType=S \
            --key-schema AttributeName=pk,KeyType=HASH AttributeName=sk,KeyType=RANGE \
            --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
            | jq -r '.TableDescription.TableArn')
        echo "Table created. ARN: $TABLE_ARN" >&2
    else
        echo "Table already exists. ARN: $TABLE_ARN" >&2
    fi

    echo $TABLE_ARN
}

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
TABLE_1_ARN=$(get_or_create_table_arn "table-1")

# Get the ARN of the queue using the queue URL
URL_SQS_ARN=$(get_or_create_queue_arn "url-to-crawl")
RAW_EVENTS_SQS_ARN=$(get_or_create_queue_arn "raw-events-queue")

# create s3 bucket if it doesn't already exist
MAIN_BUCKET_NAME="mn-revisor-chat-dev"
if awslocal s3api head-bucket --bucket $MAIN_BUCKET_NAME 2>/dev/null; then
    echo "Bucket $MAIN_BUCKET_NAME already exists."
else
    echo "Bucket $MAIN_BUCKET_NAME does not exist. Creating bucket..."
    awslocal s3 mb "s3://${MAIN_BUCKET_NAME}/"
    echo "Bucket created."
fi

# check if Sinch variables are empty
if [ -z "$SINCH_API_TOKEN" ] || [ -z "$SINCH_SERVICE_ID" ] || [ -z "$SINCH_VIRTUAL_PHONE_NUMBER" ]; then
    echo "Error: Sinch API token, service ID, and virtual phone number must be set."
    exit 1
fi

# echo the settings
echo
echo
echo "MAIN_BUCKET_NAME=\"$MAIN_BUCKET_NAME\""
echo "CHUNK_PATH_PREFIX=\"chunk\""
echo "LOCAL_ENDPOINT=\"http://localhost:4566/\""
echo "RAW_EVENTS_SQS_ARN=\"$RAW_EVENTS_SQS_ARN\""
echo "RAW_PATH_PREFIX=\"raw\""
echo "TABLE_1_ARN=\"$TABLE_1_ARN\""
echo "URL_SQS_ARN=\"$URL_SQS_ARN\""
echo "EMBEDDING_MODEL_ID=\"$EMBEDDING_MODEL_ID\""
echo "OPENSEARCH_USERNAME=\"$OPENSEARCH_USERNAME\""
echo "OPENSEARCH_PASSWORD=\"$OPENSEARCH_PASSWORD\""
echo "OPENSEARCH_DOMAIN=\"$OPENSEARCH_DOMAIN\""
echo "DO_ALLOW_OPENSEARCH_INSECURE=\"1\""
echo "OPENSEARCH_INDEX_NAME=\"$OPENSEARCH_INDEX_NAME\""
echo "SINCH_API_TOKEN=\"$SINCH_API_TOKEN\""
echo "SINCH_SERVICE_ID=\"$SINCH_SERVICE_ID\""
echo "SINCH_VIRTUAL_PHONE_NUMBER=\"$SINCH_VIRTUAL_PHONE_NUMBER\""

if [ "$OVERWRITE_SETTINGS" = true ]; then
    write_settings
else
    # prompt the user if they would like to write current settings to settings.env
    echo
    read -p "Would you like to write the current settings to settings.env? (y/n) " REPLY
    echo
    if [[ "$REPLY" == "y" || "$REPLY" == "Y" ]]; then
        write_settings
    else
        echo "Settings not written to ./settings.env"
    fi
fi
