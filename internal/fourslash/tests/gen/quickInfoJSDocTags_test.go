package fourslash_test

import (
	"testing"

	"github.com/microsoft/typescript-go/internal/fourslash"
	"github.com/microsoft/typescript-go/internal/testutil"
)

func TestQuickInfoJSDocTags(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `/**
 * This is class Foo.
 * @mytag comment1 comment2
 */
class Foo {
    /**
     * This is the constructor.
     * @myjsdoctag this is a comment
     */
    constructor(value: number) {}
    /**
     * method1 documentation
     * @mytag comment1 comment2
     */
    static method1() {}
    /**
     * @mytag
     */
    method2() {}
    /**
     * @mytag comment1 comment2
     */
    property1: string;
    /**
     * @mytag1 some comments
     * some more comments about mytag1
     * @mytag2
     * here all the comments are on a new line
     * @mytag3
     * @mytag
     */
    property2: number;
    /**
     * @returns {number} a value
     */
    method3(): number { return 3; }
    /**
     * @param {string} foo A value.
     * @returns {number} Another value
     * @mytag
     */
    method4(foo: string): number { return 3; }
    /** @mytag */
    method5() {}
    /** method documentation
     *  @mytag a JSDoc tag
     */
    newMethod() {}
}
var foo = new /*1*/Foo(/*10*/4);
/*2*/Foo./*3*/method1(/*11*/);
foo./*4*/method2(/*12*/);
foo./*5*/method3(/*13*/);
foo./*6*/method4();
foo./*7*/property1;
foo./*8*/property2;
foo./*9*/method5();
foo.newMet/*14*/`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineHover(t)
}
