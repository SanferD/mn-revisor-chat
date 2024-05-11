import * as iam from "aws-cdk-lib/aws-iam";
import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as sqs from "aws-cdk-lib/aws-sqs";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";
import { KiB, TTL_ATTRIBUTE } from "./constants";
import { TempLogGroup } from "../constructs/temp-log-group";
import * as helpers from "./helpers";
import { IamResource } from "aws-cdk-lib/aws-appsync";

const URL_SQS_NAME = "url-to-crawl";
const TRIGGER_CRAWLER_NAME = "trigger_crawler";

export interface CrawlerStackProps extends cdk.StackProps {
  nonce: string;
  securityGroup: ec2.SecurityGroup;
  vpc: ec2.Vpc;
  privateIsolatedSubnets: ec2.SelectedSubnets;
}

export class CrawlerStack extends cdk.Stack {
  readonly urlSqs: sqs.Queue;
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
    this.urlSqs = new sqs.Queue(this, "url-sqs", {
      queueName: `${URL_SQS_NAME}-src-${props.nonce}`,
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

    // setup Lambda to trigger crawler

    //// trigger-crawler log group
    const triggerCrawlerLambdaLogGroup = new TempLogGroup(this, "trigger-crawler-log-group");

    //// trigger-crawler Lambda
    helpers.codeBuild(TRIGGER_CRAWLER_NAME);
    const triggerCrawlerLambda = new lambda.Function(this, "trigger-crawler", {
      functionName: `trigger-crawler-${props.nonce}`,
      code: lambda.Code.fromAsset(helpers.getAssetPath(TRIGGER_CRAWLER_NAME)), // GoLang code
      handler: "HandleRequests", // handler function. Can be named anything, happens to be "Handler"
      runtime: lambda.Runtime.PROVIDED_AL2023, // recommended
      allowPublicSubnet: false, // network isolation => private subnets only
      environment: {
        URL_SQS_ARN: this.urlSqs.queueArn, // sqs arn over environment variables for easy dependency management
        TABLE_1_ARN: this.seenUrlTable.tableArn, // ddb arn over environment variables for easy dependency management
      },
      logGroup: triggerCrawlerLambdaLogGroup, // custom log group to simplify stack deletion
      memorySize: 512, // 512 MB
      reservedConcurrentExecutions: 1, // 1 concurrent execution since this is manually triggered
      retryAttempts: 0, // don't retry, error => failed execution
      securityGroups: [props.securityGroup],
      timeout: cdk.Duration.minutes(7), // fast running Lambda (2-3 minutes), 7 minutes just incase
      vpc: props.vpc, // network isolation => within VPC
      vpcSubnets: props.privateIsolatedSubnets, // network isolatoin => private isolated subnets only
    });

    //// configure trigger-crawler permissions
    this.urlSqs.grantPurge(triggerCrawlerLambda);
    this.urlSqs.grantSendMessages(triggerCrawlerLambda);
    this.seenUrlTable.grantReadWriteData(triggerCrawlerLambda);
    triggerCrawlerLambda.addToRolePolicy(
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
    const crawlerTaskDefinition = new ecs.TaskDefinition(this, "crawler-task-definition", {
      compatibility: ecs.Compatibility.FARGATE,
      cpu: "512", // 0.5vCPU cpu
      memoryMiB: "1024", // 1GB memory
      family: `crawler-task-dfn-family-${props.nonce}`,
      networkMode: ecs.NetworkMode.AWS_VPC, // only supported option for AWS Fargate
    });
  }
}
