import path from "path";
import * as process from "process";

export function getEnv(name: string): string {
  const res = process.env[name];
  if (!res) {
    throw new Error(`missing env: ${name}`);
  }
  return String(res);
}

export function resolvePathInput(untrustedInput: string, wd: string): string {
  const safeJoin = path.resolve(path.join(wd, untrustedInput));
  if (!(safeJoin + path.sep).startsWith(wd + path.sep)) {
    throw Error(`unsafe path ${safeJoin}`);
  }
  return safeJoin;
}
