import * as iam from "aws-cdk-lib/aws-iam";
import { Construct } from "constructs";

export interface LambdaRoleProps extends iam.RoleProps {}

export class LambdaRole extends iam.Role {
  constructor(scope: Construct, id: string, props?: LambdaRoleProps) {
    super(scope, id, {
      assumedBy: new iam.ServicePrincipal("lambda.amazonaws.com"),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole"),
        iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaVPCAccessExecutionRole"),
      ],
      ...props,
    });
  }
}
