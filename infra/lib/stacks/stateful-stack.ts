import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { Construct } from "constructs";
import { DualQueue } from "../constructs/dual-sqs";
import { S3Rule } from "../constructs/s3-rule";
import * as constants from "../constants";

const MAIN_BUCKET_ID = "main-bucket";
const TABLE1_ID = "table-1";
const URL_DQ_ID = "url-dq";
const RAW_EVENTS_DQ_ID = "raw-events-dq";
const TO_INDEX_DQ_ID = "to-index-dq";
const PUT_RAW_EVENTS_TO_RAW_EVENTS_DQ_RULE_ID = "put-raw-events-to-raw-events-dq-rule";
const PUT_CHUNK_EVENTS_TO_TO_INDEX_DQ_RULE_ID = "put-chunk-events-to-to-index-dq-rule";

export interface StatefulStackProps extends cdk.StackProps {}

export class StatefulStack extends cdk.Stack {
  readonly mainBucket: s3.Bucket;
  readonly table1: dynamodb.TableV2;
  readonly urlDQ: DualQueue;
  readonly rawEventsDQ: DualQueue;

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
