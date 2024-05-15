import * as appscaling from "aws-cdk-lib/aws-applicationautoscaling";
import * as cdk from "aws-cdk-lib";
import * as cloudwatch from "aws-cdk-lib/aws-cloudwatch";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecr_assets from "aws-cdk-lib/aws-ecr-assets";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as helpers from "./helpers";
import * as iam from "aws-cdk-lib/aws-iam";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { RAW_OBJECT_PREFIX, TTL_ATTRIBUTE } from "./constants";
import { TempLogGroup } from "../constructs/temp-log-group";
import { DualQueue } from "../constructs/dual-sqs";
import { ConfiguredFunction } from "../constructs/configured-lambda";

const URL_SQS_NAME = "url-to-crawl";
const TRIGGER_CRAWLER_NAME = "trigger_crawler";
const CRAWLER_NAME = "crawler";
const CRAWLER_STREAM_PREFIX = "crawler-";

export interface CrawlerStackProps extends cdk.StackProps {
  nonce: string;
  securityGroup: ec2.SecurityGroup;
  vpc: ec2.Vpc;
  privateIsolatedSubnets: ec2.SelectedSubnets;
  privateWithEgressSubnets: ec2.SelectedSubnets;
  mainBucket: s3.Bucket;
}

export class CrawlerStack extends cdk.Stack {
  readonly dualQueue: DualQueue;
  readonly seenUrlTable: dynamodb.TableV2;

  constructor(scope: Construct, id: string, props: CrawlerStackProps) {
    super(scope, id, props);
    props = { ...props };

    // setup dynamodb seen-url
    let billing = dynamodb.Billing.provisioned({
      // estimated cost ~$2
      readCapacity: dynamodb.Capacity.autoscaled({ maxCapacity: 5 }),
      writeCapacity: dynamodb.Capacity.autoscaled({ maxCapacity: 3 }),
    });
    this.seenUrlTable = new dynamodb.TableV2(this, "table-1", {
      tableName: `table-1-${props.nonce}`,
      partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING }, // generic pk to facilitate single table design, i.e. overloaded hash key
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING }, // generic sk to facilitate single table design, i.e. overloaded range key
      billing,
      deletionProtection: false, // simplify cleanup
      removalPolicy: cdk.RemovalPolicy.DESTROY, // delete table on stack deletion for easy cleanup of demo
      timeToLiveAttribute: TTL_ATTRIBUTE, // application sets TTL to reduce storage costs
    });

    // setup SQS infrastructure
    this.dualQueue = new DualQueue(this, URL_SQS_NAME, {
      name: URL_SQS_NAME,
      nonce: props.nonce,
    });

    // setup Lambda to trigger crawler
    const configuredFunction = new ConfiguredFunction(this, TRIGGER_CRAWLER_NAME, {
      environment: {
        URL_SQS_ARN: this.dualQueue.src.queueArn,
        TABLE_1_ARN: this.seenUrlTable.tableArn,
      },
      name: TRIGGER_CRAWLER_NAME,
      nonce: props.nonce,
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });

    //// configure trigger-crawler permissions
    this.dualQueue.src.grantPurge(configuredFunction);
    this.dualQueue.src.grantSendMessages(configuredFunction);
    this.seenUrlTable.grantReadWriteData(configuredFunction);
    configuredFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["sqs:ListQueues", "dynamodb:ListTables"],
        effect: iam.Effect.ALLOW,
        resources: ["*"],
      })
    );

    // setup ecs to run crawlers

    //// setup crawler cluster
    const crawlerCluster = new ecs.Cluster(this, "crawler-cluster", {
      clusterName: `crawler-cluster-${props.nonce}`,
      containerInsights: true, // enable container insights
      enableFargateCapacityProviders: true, // use Fargate for capacity management
      vpc: props.vpc, // over VPC
    });

    //// setup task definition

    ////// task definition
    const crawlerTaskDefinition = new ecs.TaskDefinition(this, "crawler-task-definition", {
      compatibility: ecs.Compatibility.FARGATE,
      cpu: "512", // 0.5vCPU cpu
      memoryMiB: "1024", // 1GB memory
      family: `crawler-task-dfn-family-${props.nonce}`,
      networkMode: ecs.NetworkMode.AWS_VPC, // only supported option for AWS Fargate
    });

    ////// configure crawler service permissions
    this.dualQueue.src.grantConsumeMessages(crawlerTaskDefinition.taskRole);
    this.seenUrlTable.grantReadWriteData(crawlerTaskDefinition.taskRole);
    props.mainBucket.grantPut(crawlerTaskDefinition.taskRole);
    crawlerTaskDefinition.addToTaskRolePolicy(
      new iam.PolicyStatement({
        actions: ["sqs:ListQueues", "dynamodb:ListTables"],
        effect: iam.Effect.ALLOW,
        resources: ["*"],
      })
    );

    ////// docker image asset
    helpers.doMakeBuildEcs(CRAWLER_NAME);
    const crawlerDockerImageAsset = new ecr_assets.DockerImageAsset(this, "crawler-docker-image-asset", {
      directory: helpers.getCodeDirPath(),
      buildArgs: {
        BINARY_PATH: helpers.getBuildAssetPathRelativeToCodeDir(CRAWLER_NAME),
      },
      file: helpers.getCmdDockerfilePathRelativeToCodeDir(CRAWLER_NAME),
    });

    ////// crawler log group
    const crawlerEcsLogGroup = new TempLogGroup(this, "crawler-ecs-log-group");

    ////// setup container definition that uses docker image asset and task definition
    new ecs.ContainerDefinition(this, "crawler-container-definition", {
      containerName: `crawler-container-definition-${props.nonce}`,
      image: ecs.ContainerImage.fromDockerImageAsset(crawlerDockerImageAsset),
      taskDefinition: crawlerTaskDefinition,
      environment: {
        BUCKET_NAME: props.mainBucket.bucketName,
        RAW_PATH_PREFIX: RAW_OBJECT_PREFIX,
        TABLE_1_ARN: this.seenUrlTable.tableArn,
        URL_SQS_ARN: this.dualQueue.src.queueArn,
      },
      logging: new ecs.AwsLogDriver({
        logGroup: crawlerEcsLogGroup,
        streamPrefix: CRAWLER_STREAM_PREFIX,
      }),
    });

    //// setup crawler fargate service
    const crawlerFargateService = new ecs.FargateService(this, "crawler-fargate-service", {
      cluster: crawlerCluster,
      taskDefinition: crawlerTaskDefinition,
      assignPublicIp: false, // network isolation => tasks over private subnets only
      desiredCount: 0, // start out with 0 tasks and autoscale out as tasks populate the sqs queue
      maxHealthyPercent: 100, // no more than 100% so as to be polite to mn-revisor-chat
      minHealthyPercent: 0, // ok if no tasks are running for a brief period during deployment
      circuitBreaker: { enable: true, rollback: true }, // rollback on failure
      deploymentController: { type: ecs.DeploymentControllerType.ECS }, // i.e. drop some tasks and replace with newer versions
      securityGroups: [props.securityGroup],
      serviceName: `crawler-fargate-service-${props.nonce}`,
      vpcSubnets: props.privateWithEgressSubnets,
    });

    //// setup crawler fargate service autoscale
    const crawlerScalableTaskCount = crawlerFargateService.autoScaleTaskCount({
      minCapacity: 0,
      maxCapacity: 6, // number of tasks <= 6 for politeness (6 tasks @ 3 RPS/task = 2 RPS)
    });
    const backlogPerTaskMetric = new cloudwatch.MathExpression({
      expression: "FILL(sqs_messages_count, 0) / (   IF(FILL(num_tasks, 0)==0, 1, num_tasks)   )",
      usingMetrics: {
        sqs_messages_count: new cloudwatch.Metric({
          namespace: "AWS/SQS",
          metricName: "ApproximateNumberOfMessagesVisible",
          dimensionsMap: { QueueName: this.dualQueue.src.queueName },
          statistic: "average",
          period: cdk.Duration.minutes(1),
        }),
        num_tasks: new cloudwatch.Metric({
          namespace: "ECS/ContainerInsights",
          metricName: "RunningTaskCount",
          dimensionsMap: {
            ClusterName: crawlerCluster.clusterName,
            ServiceName: crawlerFargateService.serviceName,
          },
          statistic: "average",
          period: cdk.Duration.minutes(1),
        }),
      },
      label: "Backlog per Task",
    });

    crawlerScalableTaskCount.scaleOnMetric("crawler-scalable-metric", {
      metric: backlogPerTaskMetric, // x >= 0 (always positive)
      scalingSteps: [
        { lower: 0, upper: 1, change: -1 }, // x == 0; then -1
        { lower: 1, change: +1 }, // 1 <= x ; then +1
      ],
      adjustmentType: appscaling.AdjustmentType.CHANGE_IN_CAPACITY,
      cooldown: cdk.Duration.minutes(5), // 4s/url/task @ 5mins*60s => >=75 urls/task before rechecking
    });
  }
}
