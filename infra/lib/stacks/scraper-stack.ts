import * as cdk from "aws-cdk-lib";
import * as eventsources from "aws-cdk-lib/aws-lambda-event-sources";
import { Construct } from "constructs";
import { ConfiguredFunction } from "../constructs/configured-lambda";
import * as constants from "../constants";
import * as helpers from "../helpers";
import { CommonStackProps } from "./common-stack-props";

export interface ScraperStackProps extends CommonStackProps {}

export class ScraperStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: ScraperStackProps) {
    super(scope, id, props);

    // trigger scraper lambda on rawEventsDQ messages
    const scraperFunction = new ConfiguredFunction(this, constants.RAW_SCRAPER_CMD, {
      environment: helpers.getEnvironment(props),
      timeout: constants.SCRAPER_TIMEOUT_DURATION,
      securityGroup: props.securityGroup,
      vpc: props.vpc,
      vpcSubnets: props.privateIsolatedSubnets,
    });
    scraperFunction.addEventSource(new eventsources.SqsEventSource(props.rawEventsDQ.src));

    //// add permissions to scraper lambda role
    props.urlDQ.src.grantSendMessages(scraperFunction);
    props.mainBucket.grantRead(scraperFunction, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    props.mainBucket.grantDelete(scraperFunction, constants.RAW_OBJECT_PREFIX_PATH_WILDCARD);
    props.mainBucket.grantPut(scraperFunction, constants.CHUNK_OBJECT_PREFIX_PATH_WILDCARD);
    scraperFunction.addToRolePolicy(helpers.getListPolicy({ queues: true }));
  }
}
