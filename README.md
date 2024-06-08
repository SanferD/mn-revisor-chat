# mn-revisor-chat

A chat app for querying Minnesota law from the MN Revisor website.

# Architecture

![architecture](./static/mn-revisors-arch.drawio.png)

## How it Works

A crawler and scraper collaborate to find, download, and parse statutes from the MN Revisor website.
These statutes are indexed in an OpenSearch vector index.
A RAG implementation is used to answer user questions.
Users text a Sinch virtual number, triggering a RAG lookup and answer generation, with responses sent back via Sinch.

## The Parts

### Data Population

1. **main-bucket**: S3 bucket storing raw webpages with **raw/** object prefix, and statute subdivisions with **chunk/** object prefix.
1. **table-1**: DynamoDB table tracking crawled URLs.
1. **url-dq**: SQS standard queue with DLQ for URLs to be crawled.
1. **crawler service**: ECS service with autoscaling, up to 6 tasks, checks **url-dq** and **table-1**, downloads and stores webpages in **s3://main-bucket/raw/**.
1. **raw-events-dq**: SQS standard queue for s3 events.
1. **scraper**: Lambda parses raw web pages, extracts URLs (sent to url-dq) and statutes (stored in **s3://main-bucket/chunk/**).
1. **to-index-dq**: SQS standard queue for s3 events.
1. **OpenSearch vector index**: holds the document embeddings along with their IDs.
1. **indexer**: Lambda gets object keys from **to-index-dq**, obtains embeddings, and stores them in OpenSearch vector index. AWS Bedrock is used to obtain Amazon Titan V2 embeddings.

To initiate data population, an operator triggers invoke-trigger-crawler Lambda, which spawns trigger-crawler ECS task to clear table-1 and url-dq, and send seed URL to url-dq.

### RAG Answerer

1. **User** sends an SMS to a Sinch virtual number.
1. **Sinch** triggers a webhook.
1. **webhook** sends a POST request to **API Gateway**.
1. **API Gateway** triggers answerer Lambda.
1. **answerer** Lambda processes the message, gets the prompt embedding, finds top k matching documents from OpenSearch, fetches documents from **s3://main-bucket/chunk/**, augments the prompt, sends it to Claude, and texts the response back to the user. AWS Bedrock is used to obtain Amazon Titan V2 embeddings and Claude for answer generation.

### Security

All components are within a VPC. The **crawler** and **answerer** are in private-with-egress subnets, while others are in private-isolated subnets. OpenSearch EC2 instances are in private-isolated subnets. Security groups allow inbound traffic only from within the VPC. IAM roles are minimally permissive for necessary operations.

# Setup

## Prerequisites:

1. Golang
1. AWS CLI
1. AWS CDK v2
1. Docker

## Steps

### Sinch Setup

1. **Create an Account**: Sign up for a [Sinch](https://www.sinch.com/) account.
1. **Create a Virtual Number**: Go to **Numbers > Overview > GET 10DLC**, search, and _Get_ a number.
1. **Create a Conversation API App**: Navigate to **Conversation API > Overview**. Create a _NEW APP_, then select it. Set up **Channels > SMS** and choose a Service Plan ID.
1. **Register a Webhook**: For testing, get a webhook URL from a test site like [webhook.site](https://webhook.site/). Then go to **Conversation API > Apps > <app created in previous step> > ADD WEBHOOK** and enter the webhook URL in the Target URL field. Under Triggers, select MESSAGE_INBOUND.
1. **Send a Test Message**: Send a test message to the Sinch virtual number. The response should appear on the [webhook.site](https://webhook.site/) page. You might need to verify your personal number under **Numbers > Verified Numbers**.

### Infrastructure Setup

Execute these instructions within the **infra** directory.

1. **Create config.yaml**: Create a config.yaml file with the following contents:
   Obtain the Sinch config values from the Sinch website.

```
azCount: 1
nonce: <enter some nonce here>
sinchApiToken:
sinchProjectId:
sinchVirtualPhoneNumber:
```

1. **Deploy Stacks**: Run the helper script `./deploy.sh` to deploy all the stacks.
1. **Add the API Gateway URL to the Webhook**: Obtain the API Gateway URL and add it as a Sinch webhook (see **Sinch Setup > Register a Webhook**). Example output:

```
Outputs:
answerer-stack-mnrevisor.ApiEndpoint = https://abcd.execute-api.us-east-1.amazonaws.com/api/v1
```

### Index Population

1. Sign in to the AWS console.
1. Navigate to the Lambda page and run the **invoke-trigger-crawler** Lambda with any event.
1. Wait until all the SQS queues are empty (i.e., crawling has completed). This should take approximately 4 hours.
1. Should also see many documents in the OpenSearch vector index (**AWS Console > OpenSearch > opensearch domain > Instance health > Cluter health > Overall health > Searchable documents**).

![searchable documents](./static/searchable-documents.png)

### Ask a Question

Once the OpenSearch Vector Index is fully populated, it's ready to answer questions. Send a test prompt to the Sinch virtual number and receive a response after a minute or so. If you've configured a testing webhook site, you should see the output on the webhook site page.
