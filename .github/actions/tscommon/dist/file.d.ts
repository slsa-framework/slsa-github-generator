/// <reference types="node" />
/// <reference types="node" />
import fs from "fs";
export declare function resolvePathInput(input: string): string;
export declare function writeFileSync(outputFn: string, data: string | Buffer): void;
export declare function mkdirSync(outputFn: string, options: fs.MakeDirectoryOptions & {
    recursive: true;
}): void;
export declare function readFileSync(inputFn: string): Buffer;
