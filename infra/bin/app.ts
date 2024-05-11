#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import * as conf from "../lib/config";
import { S3Stack } from "../lib/s3-stack";
import { CrawlerStack } from "../lib/crawler-stack";
import { VpcStack } from "../lib/vpc-stack";

function main(app: cdk.App, config: conf.Config) {
  console.log("config: ", config);

  const i = (x: string) => `${x}-${config.nonce}`;
  const s3Stack = new S3Stack(app, i("s3-stack"), { nonce: config.nonce });

  const vpcStack = new VpcStack(app, i("vpc-stack"), { nonce: config.nonce, azCount: config.azCount });

  const crawlerStack = new CrawlerStack(app, i("crawler-stack"), {
    nonce: config.nonce,
    securityGroup: vpcStack.securityGroup,
    vpc: vpcStack.vpc,
    privateIsolatedSubnets: vpcStack.privateIsolatedSubnets,
  });
}

const app = new cdk.App();
const config = conf.parseConfig();
main(app, config);
