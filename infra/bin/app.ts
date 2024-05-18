#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import * as conf from "../lib/config";
import { S3Stack } from "../lib/stacks/s3-stack";
import { CrawlerStack } from "../lib/stacks/crawler-stack";
import { VpcStack } from "../lib/stacks/vpc-stack";
import { ScraperStack } from "../lib/stacks/scraper-stack";
import { IndexerStack } from "../lib/stacks/indexer-stack";

function main(app: cdk.App, config: conf.Config) {
  console.log("config: ", config);

  const i = (x: string) => `${x}-${config.nonce}`;
  const s3Stack = new S3Stack(app, i("s3-stack"), { nonce: config.nonce });

  const vpcStack = new VpcStack(app, i("vpc-stack"), { nonce: config.nonce, azCount: config.azCount });

  const crawlerStack = new CrawlerStack(app, i("crawler-stack"), {
    mainBucket: s3Stack.mainBucket,
    nonce: config.nonce,
    privateIsolatedSubnets: vpcStack.privateIsolatedSubnets,
    privateWithEgressSubnets: vpcStack.privateWithEgressSubnets,
    securityGroup: vpcStack.securityGroup,
    vpc: vpcStack.vpc,
  });

  const scraperStack = new ScraperStack(app, i("scraper-stack"), {
    mainBucket: s3Stack.mainBucket,
    nonce: config.nonce,
    vpc: vpcStack.vpc,
    securityGroup: vpcStack.securityGroup,
    privateIsolatedSubnets: vpcStack.privateIsolatedSubnets,
    urlDualQueue: crawlerStack.dualQueue,
    triggerCrawlerFunction: crawlerStack.triggerCrawlerFunction,
  });
}

const app = new cdk.App();
const config = conf.parseConfig();
main(app, config);
