import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import { CodeContainerDefinition } from "../constructs/code-container-definition";
import { ConfiguredTaskDefinition } from "../constructs/configured-task-definition";
import { CrawlerBacklogAutoScalingService } from "../constructs/crawler-backlog-auto-scaling-service";
import { ConfiguredCluster } from "../constructs/configured-cluster";
import * as constants from "../constants";
import * as helpers from "../helpers";
import { CommonStackProps } from "./common-stack-props";

const CRAWLER_CMD = "crawler";
const CRAWLER_SERVICE_ID = "crawler-service";
const CRAWLER_CLUSTER_ID = "crawler-cluster";
const CRAWLER_TASK_DEFINITION_ID = "crawler-task-definition";

export interface CrawlerStackProps extends CommonStackProps {}

export class CrawlerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: CrawlerStackProps) {
    super(scope, id, props);
    props = { ...props };

    /* crawler cluster */
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
