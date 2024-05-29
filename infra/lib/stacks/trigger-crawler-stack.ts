import * as cdk from "aws-cdk-lib";
import { CommonStackProps } from "./common-stack-props";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import { CodeContainerDefinition } from "../constructs/code-container-definition";
import { ConfiguredCluster } from "../constructs/configured-cluster";
import { ConfiguredTaskDefinition } from "../constructs/configured-task-definition";
import * as helpers from "../helpers";

const INVOKE_TRIGGER_CRAWLER_CMD = "invoke_trigger_crawler";
const TRIGGER_CRAWLER_CLUSTER_ID = "trigger-crawler-cluster";
const TRIGGER_CRAWLER_TASK_DEFINITION_ID = "trigger-crawler-task-definition";
const TRIGGER_CRAWLER_CMD = "trigger_crawler";

export interface TriggerCrawlerStackProps extends CommonStackProps {}

export class TriggerCrawlerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: TriggerCrawlerStackProps) {
    super(scope, id, props);

    const triggerCrawlerCluster = new ConfiguredCluster(this, TRIGGER_CRAWLER_CLUSTER_ID, {
      vpc: props.vpc,
    });

    const triggerCrawlerTaskDefinition = new ConfiguredTaskDefinition(this, TRIGGER_CRAWLER_TASK_DEFINITION_ID);

    props.rawEventsDQ.src.grantPurge(triggerCrawlerTaskDefinition.taskRole);
    props.urlDQ.src.grantPurge(triggerCrawlerTaskDefinition.taskRole);
    props.urlDQ.src.grantSendMessages(triggerCrawlerTaskDefinition.taskRole);
    props.table1.grantReadWriteData(triggerCrawlerTaskDefinition.taskRole);
    triggerCrawlerTaskDefinition.addToTaskRolePolicy(helpers.getListPolicy({ queues: true, tables: true }));

    new CodeContainerDefinition(this, TRIGGER_CRAWLER_CMD, {
      taskDefinition: triggerCrawlerTaskDefinition,
      environment: helpers.getEnvironment(props),
    });

    const fn = new ConfiguredFunction(this, INVOKE_TRIGGER_CRAWLER_CMD, {
      environment: helpers.getEnvironment({
        ...props,
        triggerCrawlerCluster,
        triggerCrawlerTaskDefinition,
      }),
      timeout: cdk.Duration.minutes(3), // at most 3 minutes to start the trigger crawler ecs task
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });
    triggerCrawlerTaskDefinition.grantRun(fn);
    fn.addToRolePolicy(helpers.getListTasksPolicy(triggerCrawlerCluster));
  }
}
