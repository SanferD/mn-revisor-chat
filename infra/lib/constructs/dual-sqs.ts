import * as cdk from "aws-cdk-lib";
import * as sqs from "aws-cdk-lib/aws-sqs";
import { KiB } from "../constants";
import { Construct } from "constructs";

const DLQ_KIND = "dlq";
const SRC_KIND = "src";

export interface DualQueueProps {
  src?: sqs.QueueProps;
  dlq?: sqs.QueueProps;
}

export class DualQueue extends Construct {
  readonly src: sqs.Queue;
  readonly dlq: sqs.Queue;

  constructor(scope: Construct, id: string, props: DualQueueProps) {
    super(scope, id);

    const makeId = (kind: string): string => `${id}-${kind}`;

    // create dead letter queue
    this.dlq = new sqs.Queue(this, makeId(DLQ_KIND), {
      retentionPeriod: cdk.Duration.days(14), // retain the message for 2 weeks
      visibilityTimeout: cdk.Duration.minutes(2), // set the visibility for at most 2 minutes
      removalPolicy: cdk.RemovalPolicy.DESTROY, // destroy the resources when the stack deletes
      redriveAllowPolicy: {
        // TODO: figure out how to set this to BY_QUEUE
        redrivePermission: sqs.RedrivePermission.ALLOW_ALL,
      },
      ...props.dlq,
    });

    // create sqs url queue
    this.src = new sqs.Queue(this, makeId(SRC_KIND), {
      encryption: sqs.QueueEncryption.SQS_MANAGED, // encryption at rest (phew, sqs will manage the data encryption keys)
      dataKeyReuse: cdk.Duration.days(1), // set sqs key reuse period to 1 day to minimize KMS API calls and keep costs low
      enforceSSL: true, // encryption in transit
      maxMessageSizeBytes: 10 * KiB, // 1 KB should suffice, 10 KB just in case
      visibilityTimeout: cdk.Duration.minutes(3), // task should take 1 minute to process a request, 3 minutes just-in-case
      retentionPeriod: cdk.Duration.days(7), // 1 week to debug an error
      removalPolicy: cdk.RemovalPolicy.DESTROY, // delete SQS queue on stack deletion for easy cleanup
      redriveAllowPolicy: {
        // source queue is not DLQ
        redrivePermission: sqs.RedrivePermission.DENY_ALL,
      },
      deadLetterQueue: {
        maxReceiveCount: 2, // retry twice before moving to DLQ to accommodate transient errors
        queue: this.dlq, // the DLQ
      },
      ...props.src,
    });
  }
}
