import * as cdk from "aws-cdk-lib";
import { CommonStackProps } from "./common-stack-props";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import * as constants from "../constants";
import * as helpers from "../helpers";

export interface AnswererStackProps extends CommonStackProps {}

export class AnswererStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: AnswererStackProps) {
    super(scope, id, props);

    const fn = new ConfiguredFunction(this, constants.ANSWERER_CMD, {
      environment: helpers.getEnvironment(props),
      role: props.answererRole,
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });

    props.mainBucket.grantRead(fn);
    props.opensearchDomain.grantIndexReadWrite(constants.VECTOR_INDEX_NAME, fn);
    fn.addToRolePolicy(
      helpers.getBedrockInvokePolicy(constants.TITAN_EMBEDDING_V2_MODEL_ID, constants.CLAUDE_MODEL_ID)
    );
  }
}
