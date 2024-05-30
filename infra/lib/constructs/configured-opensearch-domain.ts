import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as opensearchservice from "aws-cdk-lib/aws-opensearchservice";
import { Construct } from "constructs";

export interface ConfiguredOpensearchDomainProps extends Partial<opensearchservice.DomainProps> {
  multiAzWithStandbyEnabled: boolean;
  dataNodeInstanceType?: string;
  securityGroups: ec2.SecurityGroup[];
  vpc: ec2.Vpc;
  vpcSubnets: ec2.SubnetSelection[];
}

export class ConfiguredOpensearchDomain extends opensearchservice.Domain {
  constructor(scope: Construct, id: string, props: ConfiguredOpensearchDomainProps) {
    super(scope, id, {
      capacity: {
        dataNodeInstanceType: props.dataNodeInstanceType ?? "r5.large.search",
        dataNodes: 1,
        multiAzWithStandbyEnabled: props.multiAzWithStandbyEnabled,
      },
      enableAutoSoftwareUpdate: true,
      enableVersionUpgrade: true,
      encryptionAtRest: { enabled: true },
      enforceHttps: true,
      nodeToNodeEncryption: true,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      version: opensearchservice.EngineVersion.OPENSEARCH_2_11,
      ...props,
    });
  }
}
