import * as constants from "./constants";
import * as path from "path";
import { execSync } from "child_process";

const validMakeTargets = ["clean", "build-ecs", "build-lambda"];
const validCmds = ["crawler", "trigger_crawler"];

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
