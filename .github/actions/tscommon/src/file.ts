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
export function resolvePathInput(input: string, write: Boolean): string {
  const wd = getGitHubWorkspace();
  const resolvedInput = path.resolve(input);

  // Allowed files for read only.
  const allowedReadFiles: string[] = [process.env.GITHUB_EVENT_PATH || ""];

  for (const allowedReadFile of allowedReadFiles) {
    if (allowedReadFile === resolvedInput) {
      if (write) {
        throw Error(`unsafe write path ${resolvedInput}`);
      }
      return resolvedInput;
    }
  }

  // Allowed directories for read and write.
  const allowedDirs: string[] = [wd, "/tmp", process.env.RUNNER_TEMP || ""];
  for (const allowedDir of allowedDirs) {
    // NOTE: we call 'resolve' to normalize the directory name.
    const resolvedAllowedDir = path.resolve(allowedDir);
    if ((resolvedInput + path.sep).startsWith(resolvedAllowedDir + path.sep)) {
      return resolvedInput;
    }
  }

  throw Error(`unsafe path ${resolvedInput}`);
}

// Safe write function.
export function safeWriteFileSync(
  outputFn: string,
  data: string | Buffer
): void {
  const safeOutputFn = resolvePathInput(outputFn, true);
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
  const safeOutputFn = resolvePathInput(outputFn, true);
  fs.mkdirSync(safeOutputFn, options);
}

// Safe read file function.
export function safeReadFileSync(inputFn: string): Buffer {
  const safeInputFn = resolvePathInput(inputFn, false);
  return fs.readFileSync(safeInputFn);
}

// Safe unlink function.
export function safeUnlinkSync(inputFn: string): void {
  const safeInputFn = resolvePathInput(inputFn, true);
  return fs.unlinkSync(safeInputFn);
}

// Safe remove directory function.
export function safeRmdirSync(
  dir: string,
  options?: fs.RmOptions | undefined
): void {
  const safeDir = resolvePathInput(dir, true);
  return fs.rmdirSync(safeDir, options);
}

// Safe exist function.
export function safeExistsSync(inputFn: string): boolean {
  const safeInputFn = resolvePathInput(inputFn, false);
  return fs.existsSync(safeInputFn);
}

// Safe readdir function.
export async function safePromises_readdir(inputFn: string): Promise<string[]> {
  const safeInputFn = resolvePathInput(inputFn, false);
  return fs.promises.readdir(safeInputFn);
}

// Safe stat function.
export async function safePromises_stat(inputFn: string): Promise<fs.Stats> {
  const safeInputFn = resolvePathInput(inputFn, true);
  return fs.promises.stat(safeInputFn);
}
