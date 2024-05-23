import * as appscaling from "aws-cdk-lib/aws-applicationautoscaling";
import * as cdk from "aws-cdk-lib";
import * as cloudwatch from "aws-cdk-lib/aws-cloudwatch";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import { Construct } from "constructs";

const CRAWLER_SCALABLE_METRIC_ID = "crawler-scalable-metric";

export interface CrawlerBacklogAutoScalingServiceProps extends cdk.StackProps {
  cluster: ecs.Cluster;
  taskDefinition: ecs.TaskDefinition;
  securityGroup: ec2.SecurityGroup;
  vpcSubnets: ec2.SubnetSelection;
  queueName: string;
}

export class CrawlerBacklogAutoScalingService extends ecs.FargateService {
  constructor(scope: Construct, id: string, props: CrawlerBacklogAutoScalingServiceProps) {
    super(scope, id, {
      cluster: props.cluster,
      taskDefinition: props.taskDefinition,
      assignPublicIp: false, // network isolation => tasks over private subnets only
      desiredCount: 0, // start out with 0 tasks and autoscale out as tasks populate the sqs queue
      maxHealthyPercent: 100, // no more than 100% so as to be polite to mn-revisor-chat
      minHealthyPercent: 0, // ok if no tasks are running for a brief period during deployment
      circuitBreaker: { enable: true, rollback: true }, // rollback on failure
      deploymentController: { type: ecs.DeploymentControllerType.ECS }, // i.e. drop some tasks and replace with newer versions
      securityGroups: [props.securityGroup],
      vpcSubnets: props.vpcSubnets,
    });

    const crawlerScalableTaskCount = this.autoScaleTaskCount({
      minCapacity: 0,
      maxCapacity: 6, // number of tasks <= 6 for politeness (6 tasks @ 3 RPS/task = 2 RPS)
    });

    const backlogPerTaskMetric = new cloudwatch.MathExpression({
      expression: "FILL(sqs_messages_count, 0) / (   IF(FILL(num_tasks, 0)==0, 1, num_tasks)   )",
      usingMetrics: {
        sqs_messages_count: new cloudwatch.Metric({
          namespace: "AWS/SQS",
          metricName: "ApproximateNumberOfMessagesVisible",
          dimensionsMap: { QueueName: props.queueName },
          statistic: "max",
          period: cdk.Duration.minutes(1),
        }),
        num_tasks: new cloudwatch.Metric({
          namespace: "ECS/ContainerInsights",
          metricName: "RunningTaskCount",
          dimensionsMap: {
            ClusterName: props.cluster.clusterName,
            ServiceName: this.serviceName,
          },
          statistic: "max",
          period: cdk.Duration.minutes(1),
        }),
      },
      label: "Backlog per Task",
    });

    crawlerScalableTaskCount.scaleOnMetric(CRAWLER_SCALABLE_METRIC_ID, {
      metric: backlogPerTaskMetric, // x >= -1 (always positive)
      scalingSteps: [
        { lower: 0, upper: 0.99, change: -1 }, // 0 <= x <= 0.99; then -1 (inclusive interval apparently...)
        { lower: 1, change: +1 }, // 1 <= x ; then +1
      ],
      adjustmentType: appscaling.AdjustmentType.CHANGE_IN_CAPACITY,
      cooldown: cdk.Duration.minutes(5), // 4s/url/task @ 5mins*60s => >=75 urls/task before rechecking
    });
  }
}
