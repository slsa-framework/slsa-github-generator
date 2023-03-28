"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.readFileSync = exports.mkdirSync = exports.writeFileSync = exports.resolvePathInput = void 0;
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
// export function sayHello(): void {
//   console.log("hi");
// }
// export function sayGoodbye(): void {
//   console.log("goodbye");
// }
// Detect directory traversal for input file.
function resolvePathInput(input) {
    const wd = process.env["GITHUB_WORKSPACE"] || process.env["PWD"] || "";
    const safeJoin = path_1.default.resolve(path_1.default.join(wd, input));
    if (!(safeJoin + path_1.default.sep).startsWith(wd + path_1.default.sep)) {
        throw Error(`unsafe path ${safeJoin}`);
    }
    return safeJoin;
}
exports.resolvePathInput = resolvePathInput;
// Safe write function.
function writeFileSync(outputFn, data) {
    const safeOutputFn = resolvePathInput(outputFn);
    fs_1.default.writeFileSync(safeOutputFn, data, {
        flag: "ax",
        mode: 0o600,
    });
}
exports.writeFileSync = writeFileSync;
// Safe mkdir function.
function mkdirSync(outputFn, options) {
    const safeOutputFn = resolvePathInput(outputFn);
    fs_1.default.mkdirSync(safeOutputFn, options);
}
exports.mkdirSync = mkdirSync;
// Safe read file function.
function readFileSync(inputFn) {
    const safeInputFn = resolvePathInput(inputFn);
    return fs_1.default.readFileSync(safeInputFn);
}
exports.readFileSync = readFileSync;
