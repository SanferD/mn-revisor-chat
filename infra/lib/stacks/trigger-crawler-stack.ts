import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import { DualQueue } from "../constructs/dual-sqs";
import * as helpers from "../helpers";

const TRIGGER_CRAWLER_ID = "trigger_crawler";

export interface TriggerCrawlerStackProps extends cdk.StackProps {
  vpc: ec2.Vpc;
  securityGroup: ec2.SecurityGroup;
  privateIsolatedSubnets: ec2.SubnetSelection;
  table1: dynamodb.TableV2;
  urlDQ: DualQueue;
  rawEventsDQ: DualQueue;
}

export class TriggerCrawlerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: TriggerCrawlerStackProps) {
    super(scope, id, props);

    // setup Lambda to trigger crawler
    const triggerCrawlerFunction = new ConfiguredFunction(this, TRIGGER_CRAWLER_ID, {
      environment: helpers.getEnvironment(props),
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });

    //// configure trigger-crawler permissions
    props.rawEventsDQ.src.grantPurge(triggerCrawlerFunction);
    props.urlDQ.src.grantPurge(triggerCrawlerFunction);
    props.urlDQ.src.grantSendMessages(triggerCrawlerFunction);
    props.table1.grantReadWriteData(triggerCrawlerFunction);
    triggerCrawlerFunction.addToRolePolicy(helpers.getListPolicy({ queues: true, tables: true }));
  }
}
