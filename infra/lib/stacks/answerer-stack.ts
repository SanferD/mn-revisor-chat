import * as cdk from "aws-cdk-lib";
import { HttpLambdaIntegration } from "aws-cdk-lib/aws-apigatewayv2-integrations";
import * as apigwv2 from "aws-cdk-lib/aws-apigatewayv2";
import { CommonStackProps } from "./common-stack-props";
import { SinchConfigProps } from "../constructs/sinch-config-props";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import * as constants from "../constants";
import * as helpers from "../helpers";

const ANSWERER_LAMBDA_INTEGRATION_ID = "answerer-lambda-integration";
const API_GATEWAY_ID = "api-gateway";

export interface AnswererStackProps extends CommonStackProps, SinchConfigProps {}

export class AnswererStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: AnswererStackProps) {
    super(scope, id, props);

    const answererFunction = new ConfiguredFunction(this, constants.ANSWERER_CMD, {
      environment: helpers.getEnvironment(props),
      role: props.answererRole,
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateWithEgressSubnets,
    });

    props.mainBucket.grantRead(answererFunction);
    props.opensearchDomain.grantIndexReadWrite(constants.VECTOR_INDEX_NAME, answererFunction);
    answererFunction.addToRolePolicy(
      helpers.getBedrockInvokePolicy(constants.TITAN_EMBEDDING_V2_MODEL_ID, constants.CLAUDE_MODEL_ID)
    );

    const answererFunctionIntegration = new HttpLambdaIntegration(ANSWERER_LAMBDA_INTEGRATION_ID, answererFunction);

    const httpApi = new apigwv2.HttpApi(this, API_GATEWAY_ID);
    httpApi.addRoutes({
      path: "/api/v1",
      methods: [apigwv2.HttpMethod.POST],
      integration: answererFunctionIntegration,
    });

    // Output the API endpoint to the console
    new cdk.CfnOutput(this, "ApiEndpoint", {
      value: httpApi.apiEndpoint,
      description: "The API endpoint for the Sinch webhook",
    });
  }
}
