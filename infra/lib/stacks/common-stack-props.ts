import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import { DualQueue } from "../constructs/dual-sqs";
import { ConfiguredOpensearchDomain } from "../constructs/configured-opensearch-domain";
import { LambdaRole } from "../constructs/lambda-role";

export interface CommonStackProps extends cdk.StackProps {
  // s3
  mainBucket: s3.Bucket;
  // ddb
  table1: dynamodb.TableV2;
  // sqs
  rawEventsDQ: DualQueue;
  urlDQ: DualQueue;
  toIndexDQ: DualQueue;
  // opensearch
  opensearchDomain: ConfiguredOpensearchDomain;
  // role
  answererRole: LambdaRole;
  indexerRole: LambdaRole;
  // vpc
  privateIsolatedSubnets: ec2.SubnetSelection;
  privateWithEgressSubnets: ec2.SubnetSelection;
  securityGroup: ec2.SecurityGroup;
  vpc: ec2.Vpc;
}
