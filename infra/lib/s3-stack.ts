import * as cdk from 'aws-cdk-lib'
import * as s3 from 'aws-cdk-lib/aws-s3'
import { Construct } from 'constructs';

export const RAW_OBJECT_PREFIX = "raw/"
export const TRANSFORMED_OBJECT_PREFIX = "transformed/"

export interface S3StackProps extends cdk.StackProps {
    nonce: string
}

export class S3Stack extends cdk.Stack {
    readonly crawlerBucket: s3.Bucket; 

    constructor(scope: Construct, id: string, props: S3StackProps) {
        super(scope, id, props)

        // create s3 bucket
        this.crawlerBucket = new s3.Bucket(this, "crawler-bucket", {
            encryption: s3.BucketEncryption.S3_MANAGED,
            enforceSSL: true,
            removalPolicy: cdk.RemovalPolicy.DESTROY, // for easy cleanup of demo
            bucketName: `crawler-bucket-${props.nonce}`
        })
    }
}