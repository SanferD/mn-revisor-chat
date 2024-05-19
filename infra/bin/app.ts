#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import * as conf from "../lib/config";
import * as stacks from "../lib/stacks";

function main(app: cdk.App, config: conf.Config) {
  console.log("config: ", config);

  const i = (x: string) => `${x}-${config.nonce}`;

  const statefulStack = new stacks.StatefulStack(app, i("stateful-stack"), {});
  const vpcStack = new stacks.VpcStack(app, i("vpc-stack"), { azCount: config.azCount });

  const commonProps = {
    securityGroup: vpcStack.securityGroup,
    vpc: vpcStack.vpc,
    privateIsolatedSubnets: vpcStack.privateIsolatedSubnets,
    privateWithEgressSubnets: vpcStack.privateWithEgressSubnets,
    mainBucket: statefulStack.mainBucket,
    table1: statefulStack.table1,
    urlDQ: statefulStack.urlDQ,
    rawEventsDQ: statefulStack.rawEventsDQ,
    toIndexDQ: statefulStack.toIndexDQ,
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
}

const app = new cdk.App();
const config = conf.parseConfig();
main(app, config);
