/// <reference types="node" />
/// <reference types="node" />
import fs from "fs";
export declare function getGitHubWorkspace(): string;
export declare function resolvePathInput(input: string): string;
export declare function safeWriteFileSync(outputFn: string, data: string | Buffer): void;
export declare function safeMkdirSync(outputFn: string, options: fs.MakeDirectoryOptions & {
    recursive: true;
}): void;
export declare function safeReadFileSync(inputFn: string): Buffer;
export declare function safeUnlinkSync(inputFn: string): void;
export declare function rmdirSync(dir: string, options?: fs.RmOptions | undefined): void;
export declare function safeExistsSync(inputFn: string): boolean;
export declare function safePromises_readdir(inputFn: string): Promise<string[]>;
export declare function safePromises_stat(inputFn: string): Promise<fs.Stats>;
