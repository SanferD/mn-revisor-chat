import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import { Construct } from "constructs";

const VPC_ID = "vpc";
const SECURITY_GROUP_ID = "security-group";
const VPC_GATEWAY_ENDPOINT_DYNAMODB_ID = "vpc-gateway-endpoint-dynamodb";
const VPCE_SQS_ID = "vpc-interface-endpoint-sqs";
const VPC_GATEWAY_ENDPOINT_S3_ID = "vpc-gateway-endpoint-s3";
const VPCE_BEDROCK_ID = "vpc-interface-endpoint-bedrock";
const VPCE_RUNTIME_ID = "vpc-interface-endpoint-bedrock-runtime";
const VPCE_ECS_ID = "vpc-interace-endpoint-ecs";
const VPCE_ECR_ID = "vpc-interface-endpoint-ecr";
const VPCE_ECR_DKR_ID = "vpc-interface-endpiont-ecr-dkr";
const VPCE_CLOUDWATCH_LOGS_ID = "vpc-interface-endpoint-cloudwatch-logs";
const VPCE_CLOUDWATCH_MONITORING_ID = "vpc-interface-endpoint-cloudwatch-monitoring";

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
    this.securityGroup.addIngressRule(
      this.securityGroup,
      ec2.Port.HTTPS,
      "allow inbound HTTPS traffic originating from within security-group/vpc"
    );

    // vpc endpoints
    //// vpc endpoint to DynamoDB
    this.vpc.addGatewayEndpoint(VPC_GATEWAY_ENDPOINT_DYNAMODB_ID, {
      service: ec2.GatewayVpcEndpointAwsService.DYNAMODB,
      subnets: [this.privateIsolatedSubnets],
    });

    //// vpc endpoint to S3
    this.vpc.addGatewayEndpoint(VPC_GATEWAY_ENDPOINT_S3_ID, {
      service: ec2.GatewayVpcEndpointAwsService.S3,
      subnets: [this.privateIsolatedSubnets],
    });

    //// vpc endpoints
    const id2service: { [key: string]: ec2.InterfaceVpcEndpointAwsService } = {};
    id2service[VPCE_SQS_ID] = ec2.InterfaceVpcEndpointAwsService.SQS;
    id2service[VPCE_BEDROCK_ID] = ec2.InterfaceVpcEndpointAwsService.BEDROCK;
    id2service[VPCE_RUNTIME_ID] = ec2.InterfaceVpcEndpointAwsService.BEDROCK_RUNTIME;
    id2service[VPCE_ECS_ID] = ec2.InterfaceVpcEndpointAwsService.ECS;
    id2service[VPCE_ECR_ID] = ec2.InterfaceVpcEndpointAwsService.ECR;
    id2service[VPCE_ECR_DKR_ID] = ec2.InterfaceVpcEndpointAwsService.ECR_DOCKER;
    id2service[VPCE_CLOUDWATCH_LOGS_ID] = ec2.InterfaceVpcEndpointAwsService.CLOUDWATCH_LOGS;
    id2service[VPCE_CLOUDWATCH_MONITORING_ID] = ec2.InterfaceVpcEndpointAwsService.CLOUDWATCH_MONITORING;

    for (const id in id2service) {
      this.vpc.addInterfaceEndpoint(id, {
        service: id2service[id],
        privateDnsEnabled: true,
        subnets: this.privateIsolatedSubnets,
      });
    }
  }
}
