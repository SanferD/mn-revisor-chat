import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as targets from "aws-cdk-lib/aws-events-targets";
import * as opensearch from "aws-cdk-lib/aws-opensearchservice";
import { Construct } from "constructs";
import { DualQueue } from "../constructs/dual-sqs";
import { S3Rule } from "../constructs/s3-rule";
import { ConfiguredOpensearchDomain } from "../constructs/configured-opensearch-domain";
import * as constants from "../constants";

const MAIN_BUCKET_ID = "main-bucket";
const OPENSEARCH_DOMAIN_ID = "opensearch_domain";
const PUT_RAW_EVENTS_TO_RAW_EVENTS_DQ_RULE_ID = "put-raw-events-to-raw-events-dq-rule";
const PUT_CHUNK_EVENTS_TO_TO_INDEX_DQ_RULE_ID = "chunk-events-to-to-index-dq-rule";
const RAW_EVENTS_DQ_ID = "raw-events-dq";
const TABLE1_ID = "table-1";
const URL_DQ_ID = "url-dq";
const TO_INDEX_DQ_ID = "chunk-dq";

export interface StatefulStackProps extends cdk.StackProps {
  azCount: number;
  privateIsolatedSubnets: ec2.SubnetSelection;
  securityGroup: ec2.SecurityGroup;
  vpc: ec2.Vpc;
}

export class StatefulStack extends cdk.Stack {
  readonly mainBucket: s3.Bucket;
  readonly table1: dynamodb.TableV2;
  readonly urlDQ: DualQueue;
  readonly rawEventsDQ: DualQueue;
  readonly toIndexDQ: DualQueue;
  readonly opensearchDomain: ConfiguredOpensearchDomain;

  constructor(scope: Construct, id: string, props: StatefulStackProps) {
    super(scope, id, props);

    this.mainBucket = new s3.Bucket(this, MAIN_BUCKET_ID, {
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      eventBridgeEnabled: true,
      removalPolicy: cdk.RemovalPolicy.DESTROY, // for easy cleanup of demo
      lifecycleRules: [{ expiration: cdk.Duration.days(20) }],
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
    const isMultiAz = props.azCount > 1;
    this.opensearchDomain = new ConfiguredOpensearchDomain(this, OPENSEARCH_DOMAIN_ID, {
      multiAzWithStandbyEnabled: isMultiAz,
      securityGroups: [props.securityGroup],
      vpc: props.vpc,
      vpcSubnets: [props.privateIsolatedSubnets],
    });

    /* queues for data transformation along with triggers */
    this.urlDQ = new DualQueue(this, URL_DQ_ID, {});
    this.rawEventsDQ = new DualQueue(this, RAW_EVENTS_DQ_ID, {
      src: {
        visibilityTimeout: constants.SCRAPER_TIMEOUT_DURATION,
      },
    });
    this.toIndexDQ = new DualQueue(this, TO_INDEX_DQ_ID, {
      src: {
        visibilityTimeout: constants.INDEXER_TIMEOUT_DURATION,
      },
    });

    // send PutObject events over s3://main-bucket/raw/* to the raw-events queue
    new S3Rule(this, PUT_RAW_EVENTS_TO_RAW_EVENTS_DQ_RULE_ID, {
      bucket: this.mainBucket,
      prefix: constants.RAW_OBJECT_PREFIX_PATH,
      targets: [new targets.SqsQueue(this.rawEventsDQ.src)],
    });

    // send PutObject events over s3://main-bucket/chunk/* to the chunk queue
    new S3Rule(this, PUT_CHUNK_EVENTS_TO_TO_INDEX_DQ_RULE_ID, {
      bucket: this.mainBucket,
      prefix: constants.CHUNK_OBJECT_PREFIX_PATH,
      targets: [new targets.SqsQueue(this.toIndexDQ.src)],
    });
  }
}
