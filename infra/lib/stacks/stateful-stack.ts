import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { Construct } from "constructs";
import { bedrock } from "@cdklabs/generative-ai-cdk-constructs";
import { opensearchserverless, opensearch_vectorindex } from "@cdklabs/generative-ai-cdk-constructs";
import { DualQueue } from "../constructs/dual-sqs";
import { S3Rule } from "../constructs/s3-rule";
import * as constants from "../constants";

const DATA_SOURCE_NAME = "mn-revisor-statute-subdivisions";
const KNOWLEDGE_BASE_ID = "knowledge-base";
const MAIN_BUCKET_ID = "main-bucket";
const PUT_RAW_EVENTS_TO_RAW_EVENTS_DQ_RULE_ID = "put-raw-events-to-raw-events-dq-rule";
const RAW_EVENTS_DQ_ID = "raw-events-dq";
const SUBDIVISIONS_DATA_SOURCE_ID = "subdivisions-data-source";
const TABLE1_ID = "table-1";
const URL_DQ_ID = "url-dq";
const VECTOR_COLLECTION_ID = "vector-collection";
const VECTOR_INDEX_ID = "vector-index";

export interface StatefulStackProps extends cdk.StackProps {}

export class StatefulStack extends cdk.Stack {
  readonly mainBucket: s3.Bucket;
  readonly table1: dynamodb.TableV2;
  readonly urlDQ: DualQueue;
  readonly rawEventsDQ: DualQueue;
  readonly knowledgeBase: bedrock.KnowledgeBase;

  constructor(scope: Construct, id: string, props: StatefulStackProps) {
    super(scope, id, props);

    this.mainBucket = new s3.Bucket(this, MAIN_BUCKET_ID, {
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      eventBridgeEnabled: true,
      removalPolicy: cdk.RemovalPolicy.DESTROY, // for easy cleanup of demo
    });

    this.table1 = new dynamodb.TableV2(this, TABLE1_ID, {
      partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING }, // generic pk to facilitate single table design, i.e. overloaded hash key
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING }, // generic sk to facilitate single table design, i.e. overloaded range key
      billing: dynamodb.Billing.provisioned({
        // estimated cost ~$2
        readCapacity: dynamodb.Capacity.autoscaled({ maxCapacity: 5 }),
        writeCapacity: dynamodb.Capacity.autoscaled({ maxCapacity: 3 }),
      }),
      deletionProtection: false, // simplify cleanup
      removalPolicy: cdk.RemovalPolicy.DESTROY, // delete table on stack deletion for easy cleanup of demo
      timeToLiveAttribute: constants.TTL_ATTRIBUTE, // application sets TTL to reduce storage costs
    });

    /* opensearch */
    const vectorCollection = new opensearchserverless.VectorCollection(this, VECTOR_COLLECTION_ID);

    const vectorIndex = new opensearch_vectorindex.VectorIndex(this, VECTOR_INDEX_ID, {
      collection: vectorCollection,
      indexName: constants.SUBDIVISIONS_STATUTES_INDEX_NAME,
      vectorField: constants.SUBDIVISIONS_VECTOR,
      vectorDimensions: 1536,
      mappings: [
        {
          mappingField: "AMAZON_BEDROCK_TEXT_CHUNK",
          dataType: "text",
          filterable: true,
        },
        {
          mappingField: "AMAZON_BEDROCK_METADATA",
          dataType: "text",
          filterable: false,
        },
      ],
    });

    this.knowledgeBase = new bedrock.KnowledgeBase(this, KNOWLEDGE_BASE_ID, {
      embeddingsModel: bedrock.BedrockFoundationModel.TITAN_EMBED_TEXT_V1,
      indexName: vectorIndex.indexName,
      vectorStore: vectorCollection,
      vectorField: vectorIndex.vectorField,
      vectorIndex,
    });

    new bedrock.S3DataSource(this, SUBDIVISIONS_DATA_SOURCE_ID, {
      dataSourceName: DATA_SOURCE_NAME,
      bucket: this.mainBucket,
      inclusionPrefixes: [constants.CHUNK_OBJECT_PREFIX],
      chunkingStrategy: bedrock.ChunkingStrategy.NONE,
      knowledgeBase: this.knowledgeBase,
      overlapPercentage: 0,
    });

    /* queues for data transformation along with triggers */
    this.urlDQ = new DualQueue(this, URL_DQ_ID, {});
    this.rawEventsDQ = new DualQueue(this, RAW_EVENTS_DQ_ID, {});

    // send PutObject events over s3://main-bucket/raw/* to the raw-events queue
    new S3Rule(this, PUT_RAW_EVENTS_TO_RAW_EVENTS_DQ_RULE_ID, {
      bucket: this.mainBucket,
      prefix: constants.RAW_OBJECT_PREFIX_PATH,
      targets: [new targets.SqsQueue(this.rawEventsDQ.src)],
    });
  }
}
