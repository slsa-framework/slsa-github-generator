"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
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
exports.safePromises_stat = exports.safePromises_readdir = exports.safeExistsSync = exports.safeRmdirSync = exports.safeUnlinkSync = exports.safeReadFileSync = exports.safeMkdirSync = exports.safeWriteFileSync = exports.resolvePathInput = exports.safeFileSha256 = exports.getGitHubWorkspace = void 0;
const crypto = __importStar(require("crypto"));
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
const process_1 = __importDefault(require("process"));
// This function is for unit tests.
// We need to set the working directory to the tscommon/ directory
// instead of the GITHUB_WORKSPACE.
function getGitHubWorkspace() {
    const wdt = process_1.default.env.UNIT_TESTS_WD || "";
    if (wdt) {
        return wdt;
    }
    return process_1.default.env.GITHUB_WORKSPACE || "";
}
exports.getGitHubWorkspace = getGitHubWorkspace;
// safeFileSha256 returns the hex-formatted sha256 sum of the contents of an
// untrusted file path.
function safeFileSha256(untrustedPath) {
    if (!safeExistsSync(untrustedPath)) {
        throw new Error(`File ${untrustedPath} not present`);
    }
    const untrustedFile = safeReadFileSync(untrustedPath);
    return crypto.createHash("sha256").update(untrustedFile).digest("hex");
}
exports.safeFileSha256 = safeFileSha256;
// Detect directory traversal for input file.
// This function is exported for unit tests only.
function resolvePathInput(input, write) {
    const wd = getGitHubWorkspace();
    const resolvedInput = path_1.default.resolve(input);
    // Allowed files for read only.
    const allowedReadFiles = [process_1.default.env.GITHUB_EVENT_PATH || ""];
    for (const allowedReadFile of allowedReadFiles) {
        if (allowedReadFile === resolvedInput) {
            if (write) {
                throw Error(`unsafe write path ${resolvedInput}`);
            }
            return resolvedInput;
        }
    }
    // Allowed directories for read and write.
    const allowedDirs = [wd, "/tmp", process_1.default.env.RUNNER_TEMP || ""];
    for (const allowedDir of allowedDirs) {
        // NOTE: we call 'resolve' to normalize the directory name.
        const resolvedAllowedDir = path_1.default.resolve(allowedDir);
        if ((resolvedInput + path_1.default.sep).startsWith(resolvedAllowedDir + path_1.default.sep)) {
            return resolvedInput;
        }
    }
    throw Error(`unsafe path ${resolvedInput}`);
}
exports.resolvePathInput = resolvePathInput;
// Safe write function.
function safeWriteFileSync(outputFn, data) {
    const safeOutputFn = resolvePathInput(outputFn, true);
    // WARNING: if the call fails, the type of the error is not 'Error'.
    fs_1.default.writeFileSync(safeOutputFn, data, {
        flag: "wx",
        mode: 0o600,
    });
}
exports.safeWriteFileSync = safeWriteFileSync;
// Safe mkdir function.
function safeMkdirSync(outputFn, options) {
    const safeOutputFn = resolvePathInput(outputFn, true);
    fs_1.default.mkdirSync(safeOutputFn, options);
}
exports.safeMkdirSync = safeMkdirSync;
// Safe read file function.
function safeReadFileSync(inputFn) {
    const safeInputFn = resolvePathInput(inputFn, false);
    return fs_1.default.readFileSync(safeInputFn);
}
exports.safeReadFileSync = safeReadFileSync;
// Safe unlink function.
function safeUnlinkSync(inputFn) {
    const safeInputFn = resolvePathInput(inputFn, true);
    return fs_1.default.unlinkSync(safeInputFn);
}
exports.safeUnlinkSync = safeUnlinkSync;
// Safe remove directory function.
function safeRmdirSync(dir, options) {
    const safeDir = resolvePathInput(dir, true);
    return fs_1.default.rmdirSync(safeDir, options);
}
exports.safeRmdirSync = safeRmdirSync;
// Safe exist function.
function safeExistsSync(inputFn) {
    const safeInputFn = resolvePathInput(inputFn, false);
    return fs_1.default.existsSync(safeInputFn);
}
exports.safeExistsSync = safeExistsSync;
// Safe readdir function.
function safePromises_readdir(inputFn) {
    return __awaiter(this, void 0, void 0, function* () {
        const safeInputFn = resolvePathInput(inputFn, false);
        return fs_1.default.promises.readdir(safeInputFn);
    });
}
exports.safePromises_readdir = safePromises_readdir;
// Safe stat function.
function safePromises_stat(inputFn) {
    return __awaiter(this, void 0, void 0, function* () {
        const safeInputFn = resolvePathInput(inputFn, true);
        return fs_1.default.promises.stat(safeInputFn);
    });
}
exports.safePromises_stat = safePromises_stat;
