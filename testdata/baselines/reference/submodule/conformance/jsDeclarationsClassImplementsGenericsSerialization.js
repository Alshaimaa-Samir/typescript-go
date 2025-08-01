//// [tests/cases/conformance/jsdoc/declarations/jsDeclarationsClassImplementsGenericsSerialization.ts] ////

//// [interface.ts]
export interface Encoder<T> {
    encode(value: T): Uint8Array
}
//// [lib.js]
/**
 * @template T
 * @implements {IEncoder<T>}
 */
export class Encoder {
    /**
     * @param {T} value 
     */
    encode(value) {
        return new Uint8Array(0)
    }
}


/**
 * @template T
 * @typedef {import('./interface').Encoder<T>} IEncoder
 */

//// [interface.js]
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
//// [lib.js]
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Encoder = void 0;
/**
 * @template T
 * @implements {IEncoder<T>}
 */
class Encoder {
    /**
     * @param {T} value
     */
    encode(value) {
        return new Uint8Array(0);
    }
}
exports.Encoder = Encoder;


//// [interface.d.ts]
export interface Encoder<T> {
    encode(value: T): Uint8Array;
}
//// [lib.d.ts]
/**
 * @template T
 * @implements {IEncoder<T>}
 */
export declare class Encoder<T> implements IEncoder<T> {
    /**
     * @param {T} value
     */
    encode(value: T): Uint8Array<ArrayBuffer>;
}
export type IEncoder<T> = import('./interface').Encoder<T>;
