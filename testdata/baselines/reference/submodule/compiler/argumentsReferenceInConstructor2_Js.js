//// [tests/cases/compiler/argumentsReferenceInConstructor2_Js.ts] ////

//// [a.js]
class A {
	/**
	 * Constructor
	 *
	 * @param {object} [foo={}]
	 */
	constructor(foo = {}) {
		/**
		 * @type object
		 */
		this["arguments"] = foo;
	}
}




//// [a.d.ts]
declare class A {
    /**
     * Constructor
     *
     * @param {object} [foo={}]
     */
    constructor(foo?: object);
}
