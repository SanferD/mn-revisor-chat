import * as cdk from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { Construct } from "constructs";
import * as constants from "../constants";
import { S3Rule } from "../constructs/s3-rule";
import { DualQueue } from "../constructs/dual-sqs";

export interface IndexerStackProps extends cdk.StackProps {
  mainBucket: s3.Bucket;
  toIndexDQ: DualQueue;
}

export class IndexerStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: IndexerStackProps) {
    super(scope, id, props);
  }
}
