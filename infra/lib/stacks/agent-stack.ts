import * as cdk from "aws-cdk-lib";
import * as constants from "../constants";
import * as stacks from "../stacks";
import { bedrock } from "@cdklabs/generative-ai-cdk-constructs";
import { Construct } from "constructs";

const AGENT_ID = "agent";

export interface AgentStackProps extends stacks.CommonStackProps {
  knowledgeBase: bedrock.KnowledgeBase;
}

export class AgentStack extends cdk.Stack {
  readonly agent: bedrock.Agent;
  constructor(scope: Construct, id: string, props: AgentStackProps) {
    super(scope, id, props);

    this.agent = new bedrock.Agent(this, AGENT_ID, {
      aliasName: constants.BEDROCK_AGENT_NAME,
      foundationModel: bedrock.BedrockFoundationModel.ANTHROPIC_CLAUDE_V2,
      instruction: constants.MN_STATUTE_AGENT_INSTRUCTION,
      knowledgeBases: [props.knowledgeBase],
    });
  }
}
