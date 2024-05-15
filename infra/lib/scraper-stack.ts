import * as cdk from "aws-cdk-lib";
import * as events from "aws-cdk-lib/aws-events";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { Construct } from "constructs";
import { DualQueue } from "../constructs/dual-sqs";
import { RAW_OBJECT_PREFIX } from "./constants";

const RAW_EVENTS_QUEUE_NAME = "raw-events-queue";

export interface ScraperStackProps extends cdk.StackProps {
  mainBucket: s3.Bucket;
  nonce: string;
}

export class ScraperStack extends cdk.Stack {
  rawEventsQueue: DualQueue;
  constructor(scope: Construct, id: string, props: ScraperStackProps) {
    super(scope, id, props);

    // queue source
    this.rawEventsQueue = new DualQueue(this, RAW_EVENTS_QUEUE_NAME, {
      name: RAW_EVENTS_QUEUE_NAME,
      nonce: props.nonce,
    });

    // send PutObject events over s3://main-bucket/raw/* to the raw events queue
    new events.Rule(this, "main-to-raw-events-sqs-rule", {
      eventPattern: {
        source: ["aws.s3"],
        detailType: ["AWS API call"],
        detail: {
          eventName: ["PutObject"],
          requestParameters: {
            bucketName: [props.mainBucket.bucketName],
            key: [{ prefix: RAW_OBJECT_PREFIX }],
          },
        },
      },
      targets: [new targets.SqsQueue(this.rawEventsQueue.src)],
    });
  }
}
