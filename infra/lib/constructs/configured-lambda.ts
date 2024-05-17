import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";
import { TempLogGroup } from "./temp-log-group";
import * as helpers from "../helpers";

const HANDLE_REQUESTS = "HandleRequests";

export interface ConfiguredFunctionProps extends Partial<lambda.FunctionProps> {
  environment: { [key: string]: string } | undefined;
  name: string;
  nonce: string;
  securityGroup: ec2.SecurityGroup;
  vpc: ec2.Vpc;
  vpcSubnets: ec2.SubnetSelection;
}

export class ConfiguredFunction extends lambda.Function {
  constructor(scope: Construct, id: string, props: ConfiguredFunctionProps) {
    helpers.doMakeBuildLambda(props.name);
    const functionName = `${props.name}-${props.nonce}`;
    super(scope, id, {
      functionName,
      code: lambda.Code.fromAsset(helpers.getLambdaBuildAssetPath(props.name)), // GoLang code
      handler: HANDLE_REQUESTS, // handler function. Can be named anything, happens to be "Handler"
      runtime: lambda.Runtime.PROVIDED_AL2023, // recommended
      allowPublicSubnet: false, // network isolation => private subnets only
      logGroup: new TempLogGroup(scope, `${functionName}-log-group`), // custom log group to simplify stack deletion
      memorySize: 512, // 512 MB
      reservedConcurrentExecutions: 1, // 1 concurrent execution since this is manually triggered
      retryAttempts: 0, // don't retry, error => failed execution
      timeout: cdk.Duration.minutes(7), // fast running Lambda (2-3 minutes), 7 minutes just incase
      ...props,
    });
  }
}
