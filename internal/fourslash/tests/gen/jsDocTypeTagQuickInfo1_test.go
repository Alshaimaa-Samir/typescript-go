package fourslash_test

import (
	"testing"

	"github.com/microsoft/typescript-go/internal/fourslash"
	"github.com/microsoft/typescript-go/internal/testutil"
)

func TestJsDocTypeTagQuickInfo1(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @allowJs: true
// @Filename: jsDocTypeTag1.js
/** @type {String} */
var /*1*/S;
/** @type {Number} */
var /*2*/N;
/** @type {Boolean} */
var /*3*/B;
/** @type {Void} */
var /*4*/V;
/** @type {Undefined} */
var /*5*/U;
/** @type {Null} */
var /*6*/Nl;
/** @type {Array} */
var /*7*/A;
/** @type {Promise} */
var /*8*/P;
/** @type {Object} */
var /*9*/Obj;
/** @type {Function} */
var /*10*/Func;
/** @type {*} */
var /*11*/AnyType;
/** @type {?} */
var /*12*/QType;
/** @type {String|Number} */
var /*13*/SOrN;`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineHover(t)
}
