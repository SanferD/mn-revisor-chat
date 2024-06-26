#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import * as conf from "../lib/config";
import * as stacks from "../lib/stacks";
import { SinchConfigProps } from "../lib/constructs/sinch-config-props";

function main(app: cdk.App, config: conf.Config) {
  console.log("config: ", config);

  const i = (x: string) => `${x}-${config.nonce}`;

  const vpcStack = new stacks.VpcStack(app, i("vpc-stack"), { azCount: config.azCount });

  const statefulStack = new stacks.StatefulStack(app, i("stateful-stack"), {
    azCount: config.azCount,
    privateIsolatedSubnets: vpcStack.privateIsolatedSubnets,
    securityGroup: vpcStack.securityGroup,
    vpc: vpcStack.vpc,
  });

  const commonProps: stacks.CommonStackProps = {
    answererRole: statefulStack.answererRole,
    indexerRole: statefulStack.indexerRole,
    mainBucket: statefulStack.mainBucket,
    opensearchDomain: statefulStack.opensearchDomain,
    privateIsolatedSubnets: vpcStack.privateIsolatedSubnets,
    privateWithEgressSubnets: vpcStack.privateWithEgressSubnets,
    rawEventsDQ: statefulStack.rawEventsDQ,
    securityGroup: vpcStack.securityGroup,
    table1: statefulStack.table1,
    toIndexDQ: statefulStack.toIndexDQ,
    urlDQ: statefulStack.urlDQ,
    vpc: vpcStack.vpc,
  };

  const sinchConfigProps: SinchConfigProps = {
    ...config,
  };

  new stacks.TriggerCrawlerStack(app, i("trigger-crawler-stack"), {
    ...commonProps,
  });

  new stacks.CrawlerStack(app, i("crawler-stack"), {
    ...commonProps,
  });

  new stacks.ScraperStack(app, i("scraper-stack"), {
    ...commonProps,
  });

  new stacks.IndexerStack(app, i("indexer-stack"), {
    ...commonProps,
  });

  new stacks.AnswererStack(app, i("answerer-stack"), {
    ...commonProps,
    ...sinchConfigProps,
  });
}

const app = new cdk.App();
const config = conf.parseConfig();
main(app, config);
