import * as ecs from "aws-cdk-lib/aws-ecs";
import { Construct } from "constructs";

export interface ConfiguredClusterProps extends Partial<ecs.ClusterProps> {}

export class ConfiguredCluster extends ecs.Cluster {
  constructor(scope: Construct, id: string, props: ConfiguredClusterProps) {
    super(scope, id, {
      containerInsights: true, // enable container insights
      enableFargateCapacityProviders: true, // use Fargate for capacity management
      ...props,
    });
  }
}
