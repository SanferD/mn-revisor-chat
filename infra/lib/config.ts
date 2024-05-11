import * as yaml from "yaml";
import * as path from "path";
import * as fs from "fs";
import * as constants from "./constants";
import * as helpers from "./helpers";

const CONFIG_FILE_NAME = "config.yaml";

export interface Config {
  azCount: number;
  nonce: string;
}

export function parseConfig(): Config {
  const configFilePath = path.join(helpers.getRepositoryDirectory(), constants.INFRA_DIR_NAME, CONFIG_FILE_NAME);
  const configFileBuff = fs.readFileSync(configFilePath, "utf-8");
  const config = yaml.parse(configFileBuff);

  validateConfig(config);

  return {
    azCount: Number(config.azCount),
    nonce: String(config.nonce),
  };
}
function validateConfig(config: any) {
  const fields = ["azCount", "nonce"];
  const missingFields = [];
  for (var i = 0; i < fields.length; i++) {
    const field = fields[i];
    if (!(field in config)) {
      missingFields.push(field);
    }
  }
  if (missingFields.length > 0) {
    throw new Error(`${CONFIG_FILE_NAME} is missing the following fields: ${missingFields.join(", ")}`);
  }
}
