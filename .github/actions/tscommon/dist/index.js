require('./sourcemap-register.js');/******/ (() => { // webpackBootstrap
/******/ 	"use strict";
/******/ 	var __webpack_modules__ = ({

/***/ 712:
/***/ (function(__unused_webpack_module, exports, __nccwpck_require__) {


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
Object.defineProperty(exports, "__esModule", ({ value: true }));
exports.safePromises_stat = exports.safePromises_readdir = exports.safeExistsSync = exports.rmdirSync = exports.safeUnlinkSync = exports.safeReadFileSync = exports.safeReadGitHubEventFileSync = exports.safeMkdirSync = exports.safeWriteFileSync = exports.resolvePathInput = exports.getGitHubWorkspace = void 0;
const fs_1 = __importDefault(__nccwpck_require__(147));
const path_1 = __importDefault(__nccwpck_require__(17));
const process_1 = __importDefault(__nccwpck_require__(282));
// This function is for unit tests.
// We need to set the working directory to the tscommon/ directory
// instead of the GITHUB_WORKSPACE.
function getGitHubWorkspace() {
    const wdt = process_1.default.env["UNIT_TESTS_WD"] || "";
    if (wdt) {
        return wdt;
    }
    return process_1.default.env["GITHUB_WORKSPACE"] || "";
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
// Read file defined by the GitHub context,
// even if they are outside the workspace.
function safeReadGitHubEventFileSync() {
    const eventFile = process_1.default.env.GITHUB_EVENT_PATH || "";
    if (!eventFile) {
        throw Error("env GITHUB_EVENT_PATH is empty");
    }
    return fs_1.default.readFileSync(eventFile);
}
exports.safeReadGitHubEventFileSync = safeReadGitHubEventFileSync;
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


/***/ }),

/***/ 283:
/***/ (function(__unused_webpack_module, exports, __nccwpck_require__) {


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
var __exportStar = (this && this.__exportStar) || function(m, exports) {
    for (var p in m) if (p !== "default" && !Object.prototype.hasOwnProperty.call(exports, p)) __createBinding(exports, m, p);
};
Object.defineProperty(exports, "__esModule", ({ value: true }));
__exportStar(__nccwpck_require__(712), exports);


/***/ }),

/***/ 147:
/***/ ((module) => {

module.exports = require("fs");

/***/ }),

/***/ 17:
/***/ ((module) => {

module.exports = require("path");

/***/ }),

/***/ 282:
/***/ ((module) => {

module.exports = require("process");

/***/ })

/******/ 	});
/************************************************************************/
/******/ 	// The module cache
/******/ 	var __webpack_module_cache__ = {};
/******/ 	
/******/ 	// The require function
/******/ 	function __nccwpck_require__(moduleId) {
/******/ 		// Check if module is in cache
/******/ 		var cachedModule = __webpack_module_cache__[moduleId];
/******/ 		if (cachedModule !== undefined) {
/******/ 			return cachedModule.exports;
/******/ 		}
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = __webpack_module_cache__[moduleId] = {
/******/ 			// no module.id needed
/******/ 			// no module.loaded needed
/******/ 			exports: {}
/******/ 		};
/******/ 	
/******/ 		// Execute the module function
/******/ 		var threw = true;
/******/ 		try {
/******/ 			__webpack_modules__[moduleId].call(module.exports, module, module.exports, __nccwpck_require__);
/******/ 			threw = false;
/******/ 		} finally {
/******/ 			if(threw) delete __webpack_module_cache__[moduleId];
/******/ 		}
/******/ 	
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/ 	
/************************************************************************/
/******/ 	/* webpack/runtime/compat */
/******/ 	
/******/ 	if (typeof __nccwpck_require__ !== 'undefined') __nccwpck_require__.ab = __dirname + "/";
/******/ 	
/************************************************************************/
/******/ 	
/******/ 	// startup
/******/ 	// Load entry module and return exports
/******/ 	// This entry module is referenced by other modules so it can't be inlined
/******/ 	var __webpack_exports__ = __nccwpck_require__(283);
/******/ 	module.exports = __webpack_exports__;
/******/ 	
/******/ })()
;
//# sourceMappingURL=index.js.map