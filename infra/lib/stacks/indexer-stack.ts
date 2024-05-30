import * as cdk from "aws-cdk-lib";
import * as eventsources from "aws-cdk-lib/aws-lambda-event-sources";
import { CommonStackProps } from "./common-stack-props";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import * as constants from "../constants";
import * as helpers from "../helpers";

export interface IndexerStackProps extends CommonStackProps {}

export class IndexerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: IndexerStackProps) {
    super(scope, id, props);

    const fn = new ConfiguredFunction(this, constants.INDEXER_CMD, {
      environment: helpers.getEnvironment(props),
      timeout: constants.INDEXER_TIMEOUT_DURATION,
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });
    fn.addEventSource(new eventsources.SqsEventSource(props.toIndexDQ.src));

    props.toIndexDQ.src.grantConsumeMessages(fn);
    props.mainBucket.grantRead(fn);
    props.opensearchDomain.grantIndexWrite(constants.VECTOR_INDEX_NAME, fn);
    fn.addToRolePolicy(helpers.getListPolicy({ queues: true }));
    fn.addToRolePolicy(helpers.getBedrockInvokePolicy("amazon.titan-embed-text-v2:0"));
  }
}
