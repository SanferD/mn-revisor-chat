#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import * as conf from "../lib/config";
import { CrawlerStack } from "../lib/stacks/crawler-stack";
import { IndexerStack } from "../lib/stacks/indexer-stack";
import { ScraperStack } from "../lib/stacks/scraper-stack";
import { StatefulStack } from "../lib/stacks/stateful-stack";
import { VpcStack } from "../lib/stacks/vpc-stack";

function main(app: cdk.App, config: conf.Config) {
  console.log("config: ", config);

  const i = (x: string) => `${x}-${config.nonce}`;

  const statefulStack = new StatefulStack(app, i("stateful-stack"), {});
  const vpcStack = new VpcStack(app, i("vpc-stack"), { azCount: config.azCount });

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

  new CrawlerStack(app, i("crawler-stack"), {
    ...commonProps,
  });

  new ScraperStack(app, i("scraper-stack"), {
    ...commonProps,
  });

  new IndexerStack(app, i("indexer-stack"), {
    ...commonProps,
  });
}

const app = new cdk.App();
const config = conf.parseConfig();
main(app, config);
