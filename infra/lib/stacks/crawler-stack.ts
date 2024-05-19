import * as appscaling from "aws-cdk-lib/aws-applicationautoscaling";
import * as cdk from "aws-cdk-lib";
import * as cloudwatch from "aws-cdk-lib/aws-cloudwatch";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecr_assets from "aws-cdk-lib/aws-ecr-assets";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { DualQueue } from "../constructs/dual-sqs";
import { TempLogGroup } from "../constructs/temp-log-group";
import * as constants from "../constants";
import * as helpers from "../helpers";

const CRAWLER_STREAM_PREFIX = "crawler-";
const CRAWLER_CMD = "crawler";
const CRAWLER_CLUSTER_ID = "crawler-cluster";
const CRAWLER_TASK_DEFINITION_ID = "crawler-task-definition";
const CRAWLER_DOCKER_IMG_ASSET_ID = "crawler-dkr-img-asset";
const CRAWLER_CONTAINER_DFN_ID = "crawler-container-dfn";
const CRAWLER_SERVICE_ID = "crawler-service";
const CRAWLER_SCALABLE_METRIC_ID = "crawler-scalable-metric";
const CRAWLER_TEMP_LOG_GROUP_ID = `${CRAWLER_CONTAINER_DFN_ID}-log-group`;

export interface CrawlerStackProps extends cdk.StackProps {
  securityGroup: ec2.SecurityGroup;
  vpc: ec2.Vpc;
  privateIsolatedSubnets: ec2.SelectedSubnets;
  privateWithEgressSubnets: ec2.SelectedSubnets;
  mainBucket: s3.Bucket;
  table1: dynamodb.TableV2;
  urlDQ: DualQueue;
  rawEventsDQ: DualQueue;
  toIndexDQ: DualQueue;
}

export class CrawlerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: CrawlerStackProps) {
    super(scope, id, props);
    props = { ...props };

    // setup ecs to run crawlers

    //// setup crawler cluster
    const crawlerCluster = new ecs.Cluster(this, CRAWLER_CLUSTER_ID, {
      containerInsights: true, // enable container insights
      enableFargateCapacityProviders: true, // use Fargate for capacity management
      vpc: props.vpc, // over VPC
    });

    //// setup task definition

    ////// task definition
    const crawlerTaskDefinition = new ecs.TaskDefinition(this, CRAWLER_TASK_DEFINITION_ID, {
      compatibility: ecs.Compatibility.FARGATE,
      cpu: "512", // 0.5vCPU cpu
      memoryMiB: "1024", // 1GB memory
      networkMode: ecs.NetworkMode.AWS_VPC, // only supported option for AWS Fargate
    });

    ////// configure crawler service permissions
    props.urlDQ.src.grantConsumeMessages(crawlerTaskDefinition.taskRole);
    props.table1.grantReadWriteData(crawlerTaskDefinition.taskRole);
    props.mainBucket.grantPut(crawlerTaskDefinition.taskRole, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    crawlerTaskDefinition.addToTaskRolePolicy(helpers.getListPolicy({ queues: true, tables: true }));

    ////// docker image asset
    helpers.doMakeBuildEcs(CRAWLER_CMD);
    const crawlerDockerImageAsset = new ecr_assets.DockerImageAsset(this, CRAWLER_DOCKER_IMG_ASSET_ID, {
      directory: helpers.getCodeDirPath(),
      buildArgs: {
        BINARY_PATH: helpers.getBuildAssetPathRelativeToCodeDir(CRAWLER_CMD),
      },
      file: helpers.getCmdDockerfilePathRelativeToCodeDir(CRAWLER_CMD),
    });

    ////// crawler log group
    const crawlerEcsLogGroup = new TempLogGroup(this, CRAWLER_TEMP_LOG_GROUP_ID);

    ////// setup container definition that uses docker image asset and task definition
    new ecs.ContainerDefinition(this, CRAWLER_CONTAINER_DFN_ID, {
      image: ecs.ContainerImage.fromDockerImageAsset(crawlerDockerImageAsset),
      taskDefinition: crawlerTaskDefinition,
      environment: helpers.getEnvironment(props),
      logging: new ecs.AwsLogDriver({
        logGroup: crawlerEcsLogGroup,
        streamPrefix: CRAWLER_STREAM_PREFIX,
      }),
    });

    //// setup crawler fargate service
    const crawlerFargateService = new ecs.FargateService(this, CRAWLER_SERVICE_ID, {
      cluster: crawlerCluster,
      taskDefinition: crawlerTaskDefinition,
      assignPublicIp: false, // network isolation => tasks over private subnets only
      desiredCount: 0, // start out with 0 tasks and autoscale out as tasks populate the sqs queue
      maxHealthyPercent: 100, // no more than 100% so as to be polite to mn-revisor-chat
      minHealthyPercent: 0, // ok if no tasks are running for a brief period during deployment
      circuitBreaker: { enable: true, rollback: true }, // rollback on failure
      deploymentController: { type: ecs.DeploymentControllerType.ECS }, // i.e. drop some tasks and replace with newer versions
      securityGroups: [props.securityGroup],
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
          dimensionsMap: { QueueName: props.urlDQ.src.queueName },
          statistic: "max",
          period: cdk.Duration.minutes(1),
        }),
        num_tasks: new cloudwatch.Metric({
          namespace: "ECS/ContainerInsights",
          metricName: "RunningTaskCount",
          dimensionsMap: {
            ClusterName: crawlerCluster.clusterName,
            ServiceName: crawlerFargateService.serviceName,
          },
          statistic: "max",
          period: cdk.Duration.minutes(1),
        }),
      },
      label: "Backlog per Task",
    });

    crawlerScalableTaskCount.scaleOnMetric(CRAWLER_SCALABLE_METRIC_ID, {
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
