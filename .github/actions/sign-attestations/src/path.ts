import * as path from "path";

/**
 * resolvePathInput joins the path and checks that it is not outside the working
 * directory.
 */
export function resolvePathInput(input: string, wd: string): string {
  const safeJoin = path.resolve(path.join(wd, input));
  if (!(safeJoin + path.sep).startsWith(wd + path.sep)) {
    throw Error(`unsafe path ${safeJoin}`);
  }
  return safeJoin;
}
