import * as ecs from "aws-cdk-lib/aws-ecs";
import { Construct } from "constructs";

export interface ConfiguredTaskDefinitionProps extends Partial<ecs.TaskDefinitionProps> {}

export class ConfiguredTaskDefinition extends ecs.TaskDefinition {
  constructor(scope: Construct, id: string, props?: ecs.TaskDefinitionProps) {
    super(scope, id, {
      compatibility: ecs.Compatibility.FARGATE,
      cpu: "512", // 0.5vCPU cpu
      memoryMiB: "1024", // 1GB memory
      networkMode: ecs.NetworkMode.AWS_VPC, // only supported option for AWS Fargate
      ...props,
    });
  }
}
