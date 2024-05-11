import * as constants from "./constants";
import * as path from "path";
import { execSync } from "child_process";

const validMakeTargets = ["clean", "build", "build-trigger_crawler", "build-crawler"];

export function codeClean() {
  makeCode("clean");
}

export function codeBuild(name?: string) {
  let target = name === null ? "build" : `build-${name}`;
  makeCode(target);
}

export function makeCode(target: string) {
  if (!validMakeTargets.includes(target)) {
    throw new Error(`target '${target}' is not recognized`);
  }
  const rootDir = getRepositoryDirectory();
  const codeDir = path.join(rootDir, "code");
  execSync(`cd ${codeDir} && make ${target}`);
}

export function getAssetPath(target: string) {
  let rootPath = getRepositoryDirectory();
  let buildPath = path.join(rootPath, "code", ".build");
  return path.join(buildPath, target, `${target}.zip`);
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
