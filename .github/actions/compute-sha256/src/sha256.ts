import * as fs from "fs";
import * as crypto from "crypto";

/** sum generates a sha256 checksum of the given file. An error is raised if the
 * file doesn't exist. */
export function sum(untrustedPath: string): string {
  if (!fs.existsSync(untrustedPath)) {
    throw new Error(`File ${untrustedPath} not present`);
  }
  const untrustedFile = fs.readFileSync(untrustedPath);
  return crypto.createHash("sha256").update(untrustedFile).digest("hex");
}
