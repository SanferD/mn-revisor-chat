import * as cdk from "aws-cdk-lib";
export const KiB = 1024;
export const TTL_ATTRIBUTE = "ttl";
export const INFRA_DIR_NAME = "infra";
export const CHUNK_OBJECT_PREFIX = "chunk";
export const CHUNK_OBJECT_PREFIX_PATH = "chunk/";
export const CHUNK_OBJECT_PREFIX_PATH_WILDCARD = "chunk/*";
export const RAW_OBJECT_PREFIX = "raw";
export const RAW_OBJECT_PREFIX_PATH = "raw/";
export const RAW_OBJECT_PREFIX_PATH_WILDCARD = "raw/*";

export const SCRAPER_TIMEOUT_DURATION = cdk.Duration.minutes(3);
export const INDEXER_TIMEOUT_DURATION = cdk.Duration.minutes(5);

export const ANSWERER_CMD = "answerer";
export const CRAWLER_CMD = "crawler";
export const TRIGGER_CRAWLER_CMD = "trigger_crawler";
export const RAW_SCRAPER_CMD = "raw_scraper";
export const INVOKE_TRIGGER_CRAWLER_CMD = "invoke_trigger_crawler";
export const INDEXER_CMD = "indexer";
export const VALID_CMDS = [
  ANSWERER_CMD,
  CRAWLER_CMD,
  TRIGGER_CRAWLER_CMD,
  RAW_SCRAPER_CMD,
  INVOKE_TRIGGER_CRAWLER_CMD,
  INDEXER_CMD,
];

export const RAW_PATH_PREFIX_ENV_NAME = "RAW_PATH_PREFIX";
export const CHUNK_PATH_PREFIX_ENV_NAME = "CHUNK_PATH_PREFIX";
export const MAIN_BUCKET_NAME_ENV_NAME = "MAIN_BUCKET_NAME";
export const TABLE_1_ARN_ENV_NAME = "TABLE_1_ARN";
export const URL_SQS_ARN_ENV_NAME = "URL_SQS_ARN";
export const RAW_EVENTS_SQS_ARN_ENV_NAME = "RAW_EVENTS_SQS_ARN";
export const TO_INDEX_SQS_ARN_ENV_NAME = "TO_INDEX_SQS_ARN";
export const TRIGGER_CRAWLER_TASK_DEFINITION_ARN_ENV_NAME = "TRIGGER_CRAWLER_TASK_DFN_ARN";
export const TRIGGER_CRAWLER_CLUSTER_ARN_ENV_NAME = "TRIGGER_CRAWLER_CLUSTER_ARN";
export const SECURITY_GROUP_IDS_ENV_NAME = "SECURITY_GROUP_IDS";
export const PRIVATE_ISOLATED_SUBNET_IDS_ENV_NAME = "PRIVATE_ISOLATED_SUBNET_IDS";
export const BEDROCK_AGENT_NAME = "mnrevisor-bedrock-agent";

export const SUBDIVISIONS_STATUTES_INDEX_NAME = "subdivisions-vector-index";
export const SUBDIVISIONS_VECTOR = "subdivision-vector";
export const MN_STATUTE_AGENT_INSTRUCTION = `Act as an expert on Minnesota statutes. Answer the questions you are given, citing the relevant statute subdivisions. At the end of your response, include the corresponding statute subdivisions for reference. Must include citations in the response. Example:
"""
PROMPT: I want to start a construction project in the North Loop. What are some applicable laws?
RESPONSE:
The search results indicate two Minnesota laws that are relevant for construction projects in the North Loop area of Minneapolis: § 469.53 authorizes state funding for public infrastructure projects in the Regional Exchange District, which includes the North Loop. Eligible projects include construction, demolition, and site preparation. § 337.10, subd. 1, prohibits construction contracts in Minnesota from containing provisions that make the contract subject to the laws of another state.
"""
`;
export const MN_STATUTE_KNOWLEDGE_BASE_INSTRUCTION = `This knowledge base contains all Minnesota statute subdivisions. Use it to search for and cite applicable statute subdivisions.`;

export const ADMIN = "admin";
export const VECTOR_INDEX_NAME = "subdivision-knn";
export const OPENSEARCH_USERNAME_ENV_NAME = "OPENSEARCH_USERNAME";
export const OPENSEARCH_PASSWORD_ENV_NAME = "OPENSEARCH_PASSWORD";
export const OPENSEARCH_DOMAIN_ENV_NAME = "OPENSEARCH_DOMAIN";
export const OPENSEARCH_INDEX_NAME_ENV_NAME = "OPENSEARCH_INDEX_NAME";

export const TITAN_EMBEDDING_V2_MODEL_ID = "amazon.titan-embed-text-v2:0";
export const CLAUDE_MODEL_ID = "anthropic.claude-v2";
