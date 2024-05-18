import * as cdk from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { Construct } from "constructs";
import * as constants from "../constants";
import { S3Rule } from "../constructs/s3-rule";
import { DualQueue } from "../constructs/dual-sqs";

export interface IndexerStackProps extends cdk.StackProps {
  nonce: string;
  mainBucket: s3.Bucket;
}

const TO_INDEX_QUEUE_NAME = "to-index";

export class IndexerStack extends cdk.Stack {
  readonly toIndexQueue: DualQueue;
  constructor(scope: Construct, id: string, props: IndexerStackProps) {
    super(scope, id, props);

    // indexer queue source
    this.toIndexQueue = new DualQueue(this, TO_INDEX_QUEUE_NAME, {
      name: TO_INDEX_QUEUE_NAME,
      nonce: props.nonce,
    });

    // send PutObject events over s3://main-bucket/chunk/* to the indexer queue
    new S3Rule(this, "main-to-indexer-sqs-rule", {
      bucket: props.mainBucket,
      prefix: constants.CHUNK_OBJECT_PREFIX_PATH,
      targets: [new targets.SqsQueue(this.toIndexQueue.src)],
    });
  }
}
