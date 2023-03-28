import fs from "fs";
import path from "path";
import process from "process";

// This function is for unit tests.
// We need to set the working directory to the tscommon/ directory
// instead of the GITHUB_WORKSPACE.
export function getGitHubWorkspace(): string {
  const wdt = process.env["UNIT_TESTS_WD"] || "";
  if (wdt) {
    return wdt;
  }
  return process.env["GITHUB_WORKSPACE"] || "";
}

// Detect directory traversal for input file.
// This function is exported for unit tests only.
export function resolvePathInput(input: string): string {
  const wd = getGitHubWorkspace();
  const resolvedInput = path.resolve(input);
  if ((resolvedInput + path.sep).startsWith(wd + path.sep)) {
    return resolvedInput;
  }
  throw Error(`unsafe path ${resolvedInput}`);
}

// Safe write function.
export function safeWriteFileSync(
  outputFn: string,
  data: string | Buffer
): void {
  const safeOutputFn = resolvePathInput(outputFn);
  // WARNING: if the call fails, the type of the error is not 'Error'.
  fs.writeFileSync(safeOutputFn, data, {
    flag: "wx",
    mode: 0o600,
  });
}

// Safe mkdir function.
export function safeMkdirSync(
  outputFn: string,
  options: fs.MakeDirectoryOptions & { recursive: true }
): void {
  const safeOutputFn = resolvePathInput(outputFn);
  fs.mkdirSync(safeOutputFn, options);
}

// Read file defined by the GitHub context,
// even if they are outside the workspace.
export function safeReadGitHubEventFileSync(): Buffer {
  const eventFile = process.env.GITHUB_EVENT_PATH || "";
  if (!eventFile) {
    throw Error("env GITHUB_EVENT_PATH is empty");
  }
  return fs.readFileSync(eventFile);
}

// Safe read file function.
export function safeReadFileSync(inputFn: string): Buffer {
  const safeInputFn = resolvePathInput(inputFn);
  return fs.readFileSync(safeInputFn);
}

// Safe unlink function.
export function safeUnlinkSync(inputFn: string): void {
  const safeInputFn = resolvePathInput(inputFn);
  return fs.unlinkSync(safeInputFn);
}

// Safe remove directory function.
export function rmdirSync(
  dir: string,
  options?: fs.RmOptions | undefined
): void {
  const safeDir = resolvePathInput(dir);
  return fs.rmdirSync(safeDir, options);
}

// Safe exist function.
export function safeExistsSync(inputFn: string): boolean {
  const safeInputFn = resolvePathInput(inputFn);
  return fs.existsSync(safeInputFn);
}

// Safe readdir function.
export async function safePromises_readdir(inputFn: string): Promise<string[]> {
  const safeInputFn = resolvePathInput(inputFn);
  return fs.promises.readdir(safeInputFn);
}

// Safe stat function.
export async function safePromises_stat(inputFn: string): Promise<fs.Stats> {
  const safeInputFn = resolvePathInput(inputFn);
  return fs.promises.stat(safeInputFn);
}
