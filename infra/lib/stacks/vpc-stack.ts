import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import { Construct } from "constructs";

const VPC_ID = "vpc";
const SECURITY_GROUP_ID = "security-group";
const VPC_GATEWAY_ENDPOINT_DYNAMODB_ID = "vpc-gateway-endpoint-dynamodb";
const VPC_INTERFACE_ENDPOINT_SQS_ID = "vpc-interface-endpoint-sqs";
const VPC_GATEWAY_ENDPOINT_S3_ID = "vpc-gateway-endpoint-s3";
const VPC_INTERFACE_ENDPOINT_BEDROCK_ID = "vpc-interface-endpoint-bedrock";
const VPC_INTERFACE_ENDPOINT_BEDROCK_RUNTIME_ID = "vpc-interface-endpoint-bedrock-runtime";

export interface VpcStackProps extends cdk.StackProps {
  azCount: number;
}

export class VpcStack extends cdk.Stack {
  readonly vpc: ec2.Vpc;
  readonly securityGroup: ec2.SecurityGroup;
  readonly privateIsolatedSubnets: ec2.SelectedSubnets;
  readonly privateWithEgressSubnets: ec2.SelectedSubnets;

  constructor(scope: Construct, id: string, props: VpcStackProps) {
    super(scope, id, props);

    // create vpc with 1 public subnet and 2 private subnets
    this.vpc = new ec2.Vpc(this, VPC_ID, {
      ipAddresses: ec2.IpAddresses.cidr("10.0.0.0/24"), // CIDR over 24
      maxAzs: props.azCount,
      enableDnsHostnames: true, // enable DNS hostnames & DNS support => can enable private DNS names on VPC endpoints
      enableDnsSupport: true,
      subnetConfiguration: [
        {
          // public subnets => NAT gateway, internet gateway
          name: "public-subnets",
          subnetType: ec2.SubnetType.PUBLIC,
        },
        {
          // private isolated subnets => trigger-crawler lambda, vpc endpoints
          name: "private-isolated-subnets",
          subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
        },
        {
          // private egress subnets => ecs crawler service tasks
          name: "private-egress-subnets",
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
        },
      ],
    });
    this.privateIsolatedSubnets = this.vpc.selectSubnets({
      subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
    });
    this.privateWithEgressSubnets = this.vpc.selectSubnets({
      subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
    });

    // create security group which facilitates communication over HTTPs.
    this.securityGroup = new ec2.SecurityGroup(this, SECURITY_GROUP_ID, {
      vpc: this.vpc,
      allowAllOutbound: false,
    });
    this.securityGroup.addEgressRule(ec2.Peer.anyIpv4(), ec2.Port.HTTPS, "allow outbound traffic to HTTPS servers");

    // vpc endpoints
    //// vpc endpoint to DynamoDB
    this.vpc.addGatewayEndpoint(VPC_GATEWAY_ENDPOINT_DYNAMODB_ID, {
      service: ec2.GatewayVpcEndpointAwsService.DYNAMODB,
      subnets: [this.privateIsolatedSubnets],
    });

    //// vpc endpoint to SQS
    this.vpc.addInterfaceEndpoint(VPC_INTERFACE_ENDPOINT_SQS_ID, {
      service: ec2.InterfaceVpcEndpointAwsService.SQS,
      privateDnsEnabled: true,
      subnets: this.privateIsolatedSubnets,
    });

    //// vpc endpoint to S3
    this.vpc.addGatewayEndpoint(VPC_GATEWAY_ENDPOINT_S3_ID, {
      service: ec2.GatewayVpcEndpointAwsService.S3,
      subnets: [this.privateIsolatedSubnets],
    });

    //// vpc endpoint to Bedrock
    this.vpc.addInterfaceEndpoint(VPC_INTERFACE_ENDPOINT_BEDROCK_ID, {
      service: ec2.InterfaceVpcEndpointAwsService.BEDROCK,
      privateDnsEnabled: true,
      subnets: this.privateIsolatedSubnets,
    });
    this.vpc.addInterfaceEndpoint(VPC_INTERFACE_ENDPOINT_BEDROCK_RUNTIME_ID, {
      service: ec2.InterfaceVpcEndpointAwsService.BEDROCK_RUNTIME,
      privateDnsEnabled: true,
      subnets: this.privateIsolatedSubnets,
    });
  }
}
