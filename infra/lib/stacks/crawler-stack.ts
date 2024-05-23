import * as cdk from "aws-cdk-lib";
import * as cloudwatch from "aws-cdk-lib/aws-cloudwatch";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { CodeContainerDefinition } from "../constructs/code-container-definition";
import { ConfiguredTaskDefinition } from "../constructs/configured-task-definition";
import { CrawlerBacklogAutoScalingService } from "../constructs/crawler-backlog-auto-scaling-service";
import { DualQueue } from "../constructs/dual-sqs";
import * as constants from "../constants";
import * as helpers from "../helpers";
import { ConfiguredCluster } from "../constructs/configured-cluster";

const CRAWLER_CLUSTER_ID = "crawler-cluster";
const CRAWLER_CMD = "crawler";
const CRAWLER_SERVICE_ID = "crawler-service";
const CRAWLER_TASK_DEFINITION_ID = "crawler-task-definition";

export interface CrawlerStackProps extends cdk.StackProps {
  mainBucket: s3.Bucket;
  privateIsolatedSubnets: ec2.SelectedSubnets;
  privateWithEgressSubnets: ec2.SelectedSubnets;
  rawEventsDQ: DualQueue;
  securityGroup: ec2.SecurityGroup;
  table1: dynamodb.TableV2;
  urlDQ: DualQueue;
  vpc: ec2.Vpc;
}

export class CrawlerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: CrawlerStackProps) {
    super(scope, id, props);
    props = { ...props };

    const crawlerCluster = new ConfiguredCluster(this, CRAWLER_CLUSTER_ID, {
      vpc: props.vpc,
    });

    const crawlerTaskDefinition = new ConfiguredTaskDefinition(this, CRAWLER_TASK_DEFINITION_ID);
    props.urlDQ.src.grantConsumeMessages(crawlerTaskDefinition.taskRole);
    props.table1.grantReadWriteData(crawlerTaskDefinition.taskRole);
    props.mainBucket.grantPut(crawlerTaskDefinition.taskRole, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    crawlerTaskDefinition.addToTaskRolePolicy(helpers.getListPolicy({ queues: true, tables: true }));

    new CodeContainerDefinition(this, CRAWLER_CMD, {
      taskDefinition: crawlerTaskDefinition,
      environment: helpers.getEnvironment(props),
    });

    new CrawlerBacklogAutoScalingService(this, CRAWLER_SERVICE_ID, {
      cluster: crawlerCluster,
      queueName: props.urlDQ.src.queueName,
      securityGroup: props.securityGroup,
      taskDefinition: crawlerTaskDefinition,
      vpcSubnets: props.privateWithEgressSubnets,
    });
  }
}
