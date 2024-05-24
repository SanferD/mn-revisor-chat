import * as cdk from "aws-cdk-lib";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import { DualQueue } from "../constructs/dual-sqs";

export interface CommonStackProps extends cdk.StackProps {
  mainBucket: s3.Bucket;
  privateIsolatedSubnets: ec2.SubnetSelection;
  privateWithEgressSubnets: ec2.SubnetSelection;
  rawEventsDQ: DualQueue;
  securityGroup: ec2.SecurityGroup;
  table1: dynamodb.TableV2;
  urlDQ: DualQueue;
  vpc: ec2.Vpc;
}
