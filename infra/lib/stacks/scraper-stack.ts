import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as eventsources from "aws-cdk-lib/aws-lambda-event-sources";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import { DualQueue } from "../constructs/dual-sqs";
import * as constants from "../constants";
import * as helpers from "../helpers";

const RAW_SCRAPER_ID = "raw_scraper";

export interface ScraperStackProps extends cdk.StackProps {
  vpc: ec2.Vpc;
  securityGroup: ec2.SecurityGroup;
  privateIsolatedSubnets: ec2.SubnetSelection;
  mainBucket: s3.Bucket;
  urlDQ: DualQueue;
  rawEventsDQ: DualQueue;
}

export class ScraperStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: ScraperStackProps) {
    super(scope, id, props);

    // trigger scraper lambda on rawEventsDQ messages
    const scraperFunction = new ConfiguredFunction(this, RAW_SCRAPER_ID, {
      environment: helpers.getEnvironment(props),
      timeout: cdk.Duration.seconds(150),
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });
    scraperFunction.addEventSource(new eventsources.SqsEventSource(props.rawEventsDQ.src));

    //// add permissions to scraper lambda role
    props.urlDQ.src.grantSendMessages(scraperFunction);
    props.mainBucket.grantRead(scraperFunction, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    props.mainBucket.grantDelete(scraperFunction, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    props.mainBucket.grantPut(scraperFunction, constants.CHUNK_OBJECT_PREFIX_PATH_WILDCARD);
    scraperFunction.addToRolePolicy(helpers.getListPolicy({ queues: true }));
  }
}
