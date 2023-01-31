import path from "path";

// Detect directory traversal for input file.
export function resolvePathInput(input: string, wd: string): string {
  const safeJoin = path.resolve(path.join(wd, input));
  if (!(safeJoin + path.sep).startsWith(wd + path.sep)) {
    throw Error(`unsafe path ${safeJoin}`);
  }
  return safeJoin;
}
