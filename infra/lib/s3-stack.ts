import * as cdk from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";

export const MAIN_BUCKET_NAME = "main-bucket";

export interface S3StackProps extends cdk.StackProps {
  nonce: string;
}

export class S3Stack extends cdk.Stack {
  readonly mainBucket: s3.Bucket;

  constructor(scope: Construct, id: string, props: S3StackProps) {
    super(scope, id, props);

    // create s3 bucket
    this.mainBucket = new s3.Bucket(this, MAIN_BUCKET_NAME, {
      bucketName: `${MAIN_BUCKET_NAME}-${props.nonce}`,
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      removalPolicy: cdk.RemovalPolicy.DESTROY, // for easy cleanup of demo
      eventBridgeEnabled: true,
    });
  }
}
