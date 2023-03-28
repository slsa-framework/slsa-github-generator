"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.safePromises_stat = exports.safePromises_readdir = exports.safeExistsSync = exports.rmdirSync = exports.safeUnlinkSync = exports.safeReadFileSync = exports.safeMkdirSync = exports.safeWriteFileSync = exports.resolvePathInput = exports.getGitHubWorkspace = void 0;
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
// This function is for unit tests.
// We need to set the working directory to the tscommon/ directory
// instead of the GITHUB_WORKSPACE.
function getGitHubWorkspace() {
    const wdt = process.env["UNIT_TESTS_WD"] || "";
    if (wdt) {
        return wdt;
    }
    return process.env["GITHUB_WORKSPACE"] || "";
}
exports.getGitHubWorkspace = getGitHubWorkspace;
// Detect directory traversal for input file.
// This function is exported for unit tests only.
function resolvePathInput(input) {
    const wd = getGitHubWorkspace();
    const resolvedInput = path_1.default.resolve(input);
    if ((resolvedInput + path_1.default.sep).startsWith(wd + path_1.default.sep)) {
        return resolvedInput;
    }
    throw Error(`unsafe path ${resolvedInput}`);
}
exports.resolvePathInput = resolvePathInput;
// Safe write function.
function safeWriteFileSync(outputFn, data) {
    const safeOutputFn = resolvePathInput(outputFn);
    // WARNING: if the call fails, the type of the error is not 'Error'.
    fs_1.default.writeFileSync(safeOutputFn, data, {
        flag: "wx",
        mode: 0o600,
    });
}
exports.safeWriteFileSync = safeWriteFileSync;
// Safe mkdir function.
function safeMkdirSync(outputFn, options) {
    const safeOutputFn = resolvePathInput(outputFn);
    fs_1.default.mkdirSync(safeOutputFn, options);
}
exports.safeMkdirSync = safeMkdirSync;
// Safe read file function.
function safeReadFileSync(inputFn) {
    const safeInputFn = resolvePathInput(inputFn);
    return fs_1.default.readFileSync(safeInputFn);
}
exports.safeReadFileSync = safeReadFileSync;
// Safe unlink function.
function safeUnlinkSync(inputFn) {
    const safeInputFn = resolvePathInput(inputFn);
    return fs_1.default.unlinkSync(safeInputFn);
}
exports.safeUnlinkSync = safeUnlinkSync;
// Safe remove directory function.
function rmdirSync(dir, options) {
    const safeDir = resolvePathInput(dir);
    return fs_1.default.rmdirSync(safeDir, options);
}
exports.rmdirSync = rmdirSync;
// Safe exist function.
function safeExistsSync(inputFn) {
    const safeInputFn = resolvePathInput(inputFn);
    return fs_1.default.existsSync(safeInputFn);
}
exports.safeExistsSync = safeExistsSync;
// Safe readdir function.
function safePromises_readdir(inputFn) {
    return __awaiter(this, void 0, void 0, function* () {
        const safeInputFn = resolvePathInput(inputFn);
        return fs_1.default.promises.readdir(safeInputFn);
    });
}
exports.safePromises_readdir = safePromises_readdir;
// Safe stat function.
function safePromises_stat(inputFn) {
    return __awaiter(this, void 0, void 0, function* () {
        const safeInputFn = resolvePathInput(inputFn);
        return fs_1.default.promises.stat(safeInputFn);
    });
}
exports.safePromises_stat = safePromises_stat;
