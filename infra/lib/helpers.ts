import * as constants from "./constants";
import * as iam from "aws-cdk-lib/aws-iam";
import * as path from "path";
import { execSync } from "child_process";

const validMakeTargets = ["clean", "build-ecs", "build-lambda"];
const validCmds = ["crawler", "trigger_crawler", "raw_scraper"];

export function doMakeClean() {
  doMake("clean");
}

export function doMakeBuildLambda(cmd: string) {
  doMake("build-lambda", cmd);
}

export function doMakeBuildEcs(cmd: string) {
  doMake("build-ecs", cmd);
}

function doMake(target: string, cmd?: string) {
  if (!validMakeTargets.includes(target)) {
    throw new Error(`target '${target}' is not recognized`);
  }
  const rootDir = getRepositoryDirectory();
  const codeDir = path.join(rootDir, "code");
  if (cmd !== null && !validCmds.includes(cmd!)) {
    throw new Error(`cmd '${cmd} is not a valid command`);
  }
  const cmdArg = cmd === null ? "" : ` cmd=${cmd}`;
  const makeStr = `make ${target}${cmdArg}`;
  execSync(`cd ${codeDir} && ${makeStr}`);
}

export function getBuildAssetPathRelativeToCodeDir(kind: string): string {
  return `.build/${kind}/${kind}`;
}

export function getCmdDockerfilePathRelativeToCodeDir(dir: string): string {
  return `./cmd/${dir}/Dockerfile`;
}

export function getCmdDir(subdir: string): string {
  const repoDir = getRepositoryDirectory();
  return path.join(repoDir, "code", "cmd", subdir);
}

export function getEcsBuildAssetPath(target: string): string {
  return getBuildAssetPath(target, target);
}

export function getLambdaBuildAssetPath(target: string): string {
  return getBuildAssetPath(target, `${target}.zip`);
}

export function getBuildAssetPath(...targets: string[]): string {
  let buildPath: string = path.join(getCodeDirPath(), ".build");
  for (let i = 0; i < targets.length; i++) {
    buildPath = path.join(buildPath, targets[i]);
  }
  return buildPath;
}

export function getCodeDirPath(): string {
  let rootPath: string = getRepositoryDirectory();
  let codeDirPath: string = path.join(rootPath, "code");
  return codeDirPath;
}

export function getRepositoryDirectory(): string {
  let directory = __dirname;

  while (path.basename(path.resolve(directory)) !== constants.INFRA_DIR_NAME) {
    const parentDirectory = path.resolve(directory, "..");
    const isRootDir = parentDirectory === directory;
    if (isRootDir) {
      throw new Error("The 'infra' directory was not found in the path");
    }
    directory = parentDirectory;
  }

  return path.resolve(directory, ".."); // Return the parent of the 'infra' directory
}

export interface getEnvironmentProps {
  bucketName?: string;
  table1Arn?: string;
  urlSqsArn?: string;
  rawEventsSqsArn?: string;
}

export function getEnvironment(props: getEnvironmentProps): { [key: string]: string } {
  let environment: { [key: string]: string } = {};
  environment[constants.RAW_PATH_PREFIX_ENV_NAME] = constants.RAW_OBJECT_PREFIX;
  environment[constants.CHUNK_PATH_PREFIX_ENV_NAME] = constants.CHUNK_OBJECT_PREFIX;
  if (props.bucketName !== null && props.bucketName !== undefined) {
    environment[constants.BUCKET_NAME_ENV_NAME] = props.bucketName;
  }
  if (props.table1Arn !== null && props.table1Arn !== undefined) {
    environment[constants.TABLE_1_ARN_ENV_NAME] = props.table1Arn;
  }
  if (props.urlSqsArn !== null && props.urlSqsArn !== undefined) {
    environment[constants.URL_SQS_ARN_ENV_NAME] = props.urlSqsArn;
  }
  if (props.rawEventsSqsArn !== null && props.rawEventsSqsArn !== undefined) {
    environment[constants.RAW_EVENTS_SQS_ARN_ENV_NAME] = props.rawEventsSqsArn;
  }
  return environment;
}

export interface getListPolicyProps {
  queues?: boolean;
  tables?: boolean;
}

export function getListPolicy(props: getListPolicyProps): iam.PolicyStatement {
  var actions = [];
  if (props.queues ?? false) {
    actions.push("sqs:ListQueues");
  }
  if (props.tables ?? false) {
    actions.push("dynamodb:ListTables");
  }
  if (actions.length == 0) {
    throw new Error("no actions specified for getListPolicy");
  }
  return new iam.PolicyStatement({
    actions,
    effect: iam.Effect.ALLOW,
    resources: ["*"],
  });
}
