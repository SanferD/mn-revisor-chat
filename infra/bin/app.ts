#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import * as conf from "../lib/config";
import { S3Stack } from "../lib/s3-stack";

function main(app: cdk.App, config: conf.Config) {
  console.log("config: ", config);

  const i = (x: string) => `${x}-${config.nonce}`;
  const s3Stack = new S3Stack(app, i("s3-stack"), { nonce: config.nonce });
}

const app = new cdk.App();
const config = conf.parseConfig();
main(app, config);
