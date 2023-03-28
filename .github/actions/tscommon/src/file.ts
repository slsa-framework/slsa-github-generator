import fs from "fs";
import path from "path";

// export function sayHello(): void {
//   console.log("hi");
// }
// export function sayGoodbye(): void {
//   console.log("goodbye");
// }

// Detect directory traversal for input file.
export function resolvePathInput(input: string): string {
  const wd = process.env["GITHUB_WORKSPACE"] || process.env["PWD"] || "";
  const safeJoin = path.resolve(path.join(wd, input));
  if (!(safeJoin + path.sep).startsWith(wd + path.sep)) {
    throw Error(`unsafe path ${safeJoin}`);
  }
  return safeJoin;
}

// Safe write function.
export function writeFileSync(outputFn: string, data: string | Buffer): void {
  const safeOutputFn = resolvePathInput(outputFn);
  fs.writeFileSync(safeOutputFn, data, {
    flag: "ax",
    mode: 0o600,
  });
}

// Safe mkdir function.
export function mkdirSync(
  outputFn: string,
  options: fs.MakeDirectoryOptions & { recursive: true }
): void {
  const safeOutputFn = resolvePathInput(outputFn);
  fs.mkdirSync(safeOutputFn, options);
}

// Safe read file function.
export function readFileSync(inputFn: string): Buffer {
  const safeInputFn = resolvePathInput(inputFn);
  return fs.readFileSync(safeInputFn);
}
