import * as process from "process";

export function getEnv(name: string): string {
  const res = process.env[name];
  if (!res) {
    throw new Error(`missing env: ${name}`);
  }
  return String(res);
}
