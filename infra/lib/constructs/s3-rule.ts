import * as s3 from "aws-cdk-lib/aws-s3";
import * as events from "aws-cdk-lib/aws-events";
import { Construct } from "constructs";

export interface S3RuleProps extends events.RuleProps {
  bucket: s3.Bucket;
  prefix: string;
}

export class S3Rule extends events.Rule {
  constructor(scope: Construct, id: string, props: S3RuleProps) {
    super(scope, id, {
      enabled: true,
      eventPattern: {
        source: ["aws.s3"],
        detailType: ["Object Created"],
        detail: {
          bucket: {
            name: [props.bucket.bucketName],
          },
          object: {
            key: [{ prefix: props.prefix }],
          },
        },
      },
      ...props,
    });
  }
}
