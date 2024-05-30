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
      role: props.indexerRole,
      securityGroup: props.securityGroup,
      timeout: constants.INDEXER_TIMEOUT_DURATION,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });
    fn.addEventSource(new eventsources.SqsEventSource(props.toIndexDQ.src));

    props.toIndexDQ.src.grantConsumeMessages(fn);
    props.mainBucket.grantRead(fn);
    props.opensearchDomain.grantIndexWrite(constants.VECTOR_INDEX_NAME, fn);
    fn.addToRolePolicy(helpers.getListPolicy({ queues: true }));
    fn.addToRolePolicy(helpers.getBedrockInvokePolicy(constants.TITAN_EMBEDDING_V2_MODEL_ID));
  }
}
