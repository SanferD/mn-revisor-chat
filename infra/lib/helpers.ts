import * as constants from "./constants";
import * as path from "path";

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
