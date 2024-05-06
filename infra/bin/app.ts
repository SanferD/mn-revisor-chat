#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import { S3Stack } from "../lib/s3-stack";
import { CrawlerStack } from "../lib/crawler-stack";

const NONCE = "zoro";
const AZ_COUNT = 1;

function main(app: cdk.App) {
  const s3Stack = new S3Stack(app, "s3-stack", { nonce: NONCE });

  new CrawlerStack(app, "crawler-stack", {
    crawlerBucket: s3Stack.crawlerBucket,
    azCount: AZ_COUNT,
    nonce: NONCE,
  });
}

const app = new cdk.App();
main(app);
