import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as helpers from "./helpers";
import * as applicationautoscaling from "aws-cdk-lib/aws-applicationautoscaling";
import * as cloudwatch from "aws-cdk-lib/aws-cloudwatch";
import { DockerImageAsset } from "aws-cdk-lib/aws-ecr-assets";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as sqs from "aws-cdk-lib/aws-sqs";
import { TempLogGroup } from "../constructs/temp-log-group";
import { Construct } from "constructs";
import { RAW_OBJECT_PREFIX, TRANSFORMED_OBJECT_PREFIX } from "./s3-stack";
import { KiB, TTL_ATTRIBUTE } from "./constants";

const CRAWLER_STREAM_PREFIX = "crawler-";
const CRAWLER_TASK_DEFINITION_FAMILY = "crawler";
const TRIGGER_CRAWLER = "trigger_crawler";
const CRAWLER = "crawler";
const URL_SQS_NAME = "url-to-crawl";
const CRAWLER_DOCKER_CODE_DIR = "../../code/crawler/";
const TRIGGER_CRAWLER_DESCRIPTION = `
This Lambda will purge the SQS queue that contains URLs to-be scraped and 
will purge DDB table that contains URLs that have already been scraped.
It will then populate the SQS queue with an initial set of URLs to scrape.`;

export interface CrawlerStackProps extends cdk.StackProps {
  // bucket that would hold the raw data uploaded by the crawler (such as raw HTML pages)
  crawlerBucket: s3.Bucket;
  // number of azs. Default: 3
  azCount?: number;
  // string to append to the names to make them unique
  nonce: string;
}

export class CrawlerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: CrawlerStackProps) {
    super(scope, id, props);
    props = { ...props };
    props.azCount ??= 3;

    // setup dynamodb seen-url
    const seenUrlTable = new dynamodb.TableV2(this, "seen-url-table", {
      partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING }, // generic pk to facilitate single table design, i.e. overloaded hash key
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING }, // generic sk to facilitate single table design, i.e. overloaded range key
      billing: dynamodb.Billing.provisioned({
        // estimated cost ~$2
        readCapacity: dynamodb.Capacity.autoscaled({ maxCapacity: 5 }),
        writeCapacity: dynamodb.Capacity.autoscaled({ maxCapacity: 3 }),
      }),
      deletionProtection: false, // simplify cleanup
      removalPolicy: cdk.RemovalPolicy.DESTROY, // delete table on stack deletion for easy cleanup of demo
      timeToLiveAttribute: TTL_ATTRIBUTE, // application sets TTL to reduce storage costs
    });

    // setup SQS infrastructure
    //// create dead letter queue
    const urlSqsDlq = new sqs.Queue(this, "url-sqs-dlq", {
      queueName: `${URL_SQS_NAME}-dlq-${props.nonce}`,
      retentionPeriod: cdk.Duration.days(14), // retain the message for 2 weeks
      visibilityTimeout: cdk.Duration.minutes(2), // set the visibility for at most 2 minutes
      removalPolicy: cdk.RemovalPolicy.DESTROY, // destroy the resources when the stack deletes
      redriveAllowPolicy: {
        // TODO: figure out how to set this to BY_QUEUE
        redrivePermission: sqs.RedrivePermission.ALLOW_ALL,
      },
    });

    //// create sqs url-queue
    const urlSqs = new sqs.Queue(this, "url-sqs", {
      queueName: `${URL_SQS_NAME}-source-${props.nonce}`,
      encryption: sqs.QueueEncryption.SQS_MANAGED, // encryption at rest (phew, sqs will manage the data encryption keys)
      dataKeyReuse: cdk.Duration.hours(24), // set sqs key reuse period to 1 day to minimize KMS API calls and keep costs low
      enforceSSL: true, // encryption in transit
      maxMessageSizeBytes: 10 * KiB, // 1 KB should suffice, 10 KB just in case
      visibilityTimeout: cdk.Duration.minutes(3), // task should take 1 minute to process a request, 3 minutes just-in-case
      retentionPeriod: cdk.Duration.days(7), // 1 week to debug an error
      removalPolicy: cdk.RemovalPolicy.DESTROY, // delete SQS queue on stack deletion for easy cleanup of demo
      redriveAllowPolicy: {
        // source queue cannot be used as DLQ
        redrivePermission: sqs.RedrivePermission.DENY_ALL,
      },
      deadLetterQueue: {
        maxReceiveCount: 2, // retry twice before moving to DLQ to accommodate transient errors
        queue: urlSqsDlq, // the DLQ
      },
    });

    // network configuration (1/2) - vpc, subnets, NAT gateway, Internet Gateway, Security Group
    //// create vpc with 1 private subnet and 1 public subnet
    const crawlerVpc = new ec2.Vpc(this, "crawler-vpc", {
      vpcName: `crawler-vpc-${props.nonce}`,
      ipAddresses: ec2.IpAddresses.cidr("10.0.0.0/24"), // CIDR over 24
      maxAzs: props.azCount,
      enableDnsHostnames: true, // enable DNS hostnames & DNS support => can enable private DNS names on VPC endpoints
      enableDnsSupport: true,
      subnetConfiguration: [
        {
          // public subnets => NAT gateway, internet gateway
          name: "public-subnets",
          subnetType: ec2.SubnetType.PUBLIC,
        },
        {
          // private isolated subnets => trigger-crawler lambda, vpc endpoints
          name: "private-isolated-subnets",
          subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
        },
        {
          // private egress subnets => ecs crawler service tasks
          name: "private-egress-subnets",
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
        },
      ],
    });
    const privateIsolatedSubnets = crawlerVpc.selectSubnets({
      subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
    });
    const privateWithEgressSubnets = crawlerVpc.selectSubnets({
      subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
    });

    //// create security group which facilitates communication over HTTPs.
    const crawlerSecurityGroup = new ec2.SecurityGroup(this, "crawler-security-group", {
      vpc: crawlerVpc,
      allowAllOutbound: false,
    });
    crawlerSecurityGroup.addEgressRule(ec2.Peer.anyIpv4(), ec2.Port.HTTPS, "allow outbound traffic to HTTPS servers");

    // setup Lambda to trigger crawler

    //// trigger-crawler log group
    const triggerCrawlerLambdaLogGroup = new TempLogGroup(this, "trigger-crawler-log-group");

    //// trigger-crawler Lambda
    helpers.codeBuild(TRIGGER_CRAWLER);
    const triggerCrawlerLambda = new lambda.Function(this, "trigger-crawler", {
      functionName: `trigger-crawler-${props.nonce}`,
      code: lambda.Code.fromAsset(helpers.getAssetPath(TRIGGER_CRAWLER)), // GoLang code
      handler: "Handler", // handler function. Can be named anything, happens to be "Handler"
      runtime: lambda.Runtime.PROVIDED_AL2023, // recommended
      allowPublicSubnet: false, // network isolation => private subnets only
      description: TRIGGER_CRAWLER_DESCRIPTION,
      environment: {
        URL_SQS_ARN: urlSqs.queueArn, // sqs arn over environment variables for easy dependency management
        TABLE_1_ARN: seenUrlTable.tableArn, // ddb arn over environment variables for easy dependency management
      },
      logGroup: triggerCrawlerLambdaLogGroup, // custom log group to simplify stack deletion
      memorySize: 512, // 512 MB
      reservedConcurrentExecutions: 1, // 1 concurrent execution since this is manually triggered
      retryAttempts: 0, // don't retry, error => failed execution
      securityGroups: [crawlerSecurityGroup],
      timeout: cdk.Duration.minutes(7), // fast running Lambda (2-3 minutes), 7 minutes just incase
      vpc: crawlerVpc, // network isolation => within VPC
      vpcSubnets: privateIsolatedSubnets, // network isolatoin => private isolated subnets only
    });

    //// configure trigger-crawler permissions
    urlSqs.grantPurge(triggerCrawlerLambda);
    urlSqs.grantSendMessages(triggerCrawlerLambda);
    seenUrlTable.grantReadWriteData(triggerCrawlerLambda);
    return;
    // setup ecs to run crawlers

    //// setup crawler cluster
    const crawlerCluster = new ecs.Cluster(this, "crawler-cluster", {
      clusterName: `crawler-cluster-${props.nonce}`,
      containerInsights: true, // enable container insights
      enableFargateCapacityProviders: true, // use Fargate for capacity management
      vpc: crawlerVpc, // over VPC
    });

    //// setup crawler log group for crawler ecs service
    const crawlerLogGroup = new TempLogGroup(this, "crawler-log-group");

    //// setup task definition
    const crawlerTaskDefinition = new ecs.TaskDefinition(this, "crawler-task-definition", {
      compatibility: ecs.Compatibility.FARGATE,
      cpu: "512", // 0.5vCPU cpu
      memoryMiB: "1024", // 1GB memory
      family: `crawler-task-dfn-family-${props.nonce}`,
      networkMode: ecs.NetworkMode.AWS_VPC, // only supported option for AWS Fargate
    });

    //// setup container
    const asset = new DockerImageAsset(this, "crawler-docker-image-asset", {
      directory: CRAWLER_DOCKER_CODE_DIR,
    });

    new ecs.ContainerDefinition(this, "crawler-container-definition", {
      containerName: `crawler-container-definition-${props.nonce}`,
      image: ecs.ContainerImage.fromDockerImageAsset(asset),
      taskDefinition: crawlerTaskDefinition,
      environment: {
        URL_SQS_ARN: urlSqs.queueArn, // sqs arn over environment variables for easy dependency management
        SEEN_URL_DDB_ARN: seenUrlTable.tableArn, // ddb arn over environment variables for easy dependency management
      },
      logging: new ecs.AwsLogDriver({
        logGroup: crawlerLogGroup,
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
      securityGroups: [crawlerSecurityGroup],
      serviceName: `crawler-fargate-service-${props.nonce}`,
      vpcSubnets: privateWithEgressSubnets,
    });

    //// setup crawler fargate service autoscale
    const crawlerScalableTaskCount = crawlerFargateService.autoScaleTaskCount({
      minCapacity: 0,
      maxCapacity: 6, // number of tasks <= 6 for politeness (6 tasks @ 3 RPS/task = 2 RPS)
    });
    const backlogPerTaskMetric = new cloudwatch.MathExpression({
      expression: "sqs_messages_count / num_tasks",
      usingMetrics: {
        sqs_messages_count: new cloudwatch.Metric({
          namespace: "AWS/SQS",
          metricName: "ApproximateNumberOfMessagesVisible",
          dimensionsMap: { QueueName: urlSqs.queueName },
          statistic: "sum",
          period: cdk.Duration.seconds(60),
        }),
        num_tasks: new cloudwatch.Metric({
          namespace: "AWS/ECS",
          metricName: "DesiredTaskCount",
          dimensionsMap: {
            ClusterName: crawlerCluster.clusterName,
            ServiceName: crawlerFargateService.serviceName,
          },
          statistic: "average",
          period: cdk.Duration.seconds(60),
        }),
      },
      label: "Backlog per Task",
    });

    crawlerScalableTaskCount.scaleToTrackCustomMetric("crawler-scalable-metric", {
      metric: backlogPerTaskMetric,
      targetValue: 100, // 15mins=900s latency / 9 seconds per task
      scaleOutCooldown: cdk.Duration.minutes(5), // 5 minutes to register new tasks
      scaleInCooldown: cdk.Duration.minutes(5), // 5 minutes to deregister tasks
    });

    //// configure crawler service permissions
    urlSqs.grantSendMessages(crawlerTaskDefinition.taskRole);
    urlSqs.grantConsumeMessages(crawlerTaskDefinition.taskRole);
    seenUrlTable.grantReadWriteData(crawlerTaskDefinition.taskRole);
    props.crawlerBucket.grantPut(crawlerTaskDefinition.taskRole);

    // network configuration (2/2) - vpc endpoint
    //// vpc endpoint to DynamoDB
    crawlerVpc.addGatewayEndpoint("vpc-endpoint-ddb", {
      service: ec2.GatewayVpcEndpointAwsService.DYNAMODB,
      subnets: [privateIsolatedSubnets],
    });

    //// vpc endpoint to SQS
    crawlerVpc.addInterfaceEndpoint("vpc-endpoint-sqs", {
      service: ec2.InterfaceVpcEndpointAwsService.SQS,
      privateDnsEnabled: true,
      subnets: privateIsolatedSubnets,
    });
  }
}
