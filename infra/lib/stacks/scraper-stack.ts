import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as eventsources from "aws-cdk-lib/aws-lambda-event-sources";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import { DualQueue } from "../constructs/dual-sqs";
import { S3Rule } from "../constructs/s3-rule";
import * as constants from "../constants";
import * as helpers from "../helpers";

const RAW_SCRAPER_NAME = "raw_scraper";
const RAW_EVENTS_QUEUE_NAME = "raw-events-queue";

export interface ScraperStackProps extends cdk.StackProps {
  mainBucket: s3.Bucket;
  nonce: string;
  vpc: ec2.Vpc;
  securityGroup: ec2.SecurityGroup;
  privateIsolatedSubnets: ec2.SubnetSelection;
  urlDualQueue: DualQueue;
  triggerCrawlerFunction: ConfiguredFunction;
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
    new S3Rule(this, "main-to-raw-events-sqs-rule", {
      bucket: props.mainBucket,
      prefix: constants.RAW_OBJECT_PREFIX_PATH,
      targets: [new targets.SqsQueue(this.rawEventsQueue.src)],
    });

    const scraperFunction = new ConfiguredFunction(this, RAW_SCRAPER_NAME, {
      environment: helpers.getEnvironment({
        bucketName: props.mainBucket.bucketName,
        urlSqsArn: props.urlDualQueue.src.queueArn,
        rawEventsSqsArn: this.rawEventsQueue.src.queueArn,
      }),
      timeout: cdk.Duration.seconds(150),
      name: RAW_SCRAPER_NAME,
      nonce: props.nonce,
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });
    scraperFunction.addEventSource(new eventsources.SqsEventSource(this.rawEventsQueue.src));

    props.urlDualQueue.src.grantSendMessages(scraperFunction);
    this.rawEventsQueue.src.grantPurge(props.triggerCrawlerFunction);
    props.mainBucket.grantRead(scraperFunction, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    props.mainBucket.grantDelete(scraperFunction, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    props.mainBucket.grantPut(scraperFunction, constants.CHUNK_OBJECT_PREFIX_PATH_WILDCARD);
    scraperFunction.addToRolePolicy(helpers.getListPolicy({ queues: true }));
  }
}
