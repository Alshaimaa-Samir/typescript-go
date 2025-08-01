package fourslash_test

import (
	"testing"

	"github.com/microsoft/typescript-go/internal/fourslash"
	. "github.com/microsoft/typescript-go/internal/fourslash/tests/util"
	"github.com/microsoft/typescript-go/internal/testutil"
)

func TestCompletionForStringLiteral16(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `interface Foo {
    a: string;
    b: number;
    c: string;
}

declare function f1<T>(key: keyof T): T;
declare function f2<T>(a: keyof T, b: keyof T): T;

f1<Foo>("/*1*/",);
f1<Foo>("/*2*/");
f1<Foo>("/*3*/",,,);
f2<Foo>("/*4*/", "/*5*/",);
f2<Foo>("/*6*/", "/*7*/");
f2<Foo>("/*8*/", "/*9*/",,,);`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyCompletions(t, f.Markers(), &fourslash.CompletionsExpectedList{
		IsIncomplete: false,
		ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{
			CommitCharacters: &DefaultCommitCharacters,
			EditRange:        Ignored,
		},
		Items: &fourslash.CompletionsExpectedItems{
			Exact: []fourslash.CompletionsExpectedItem{
				"a",
				"b",
				"c",
			},
		},
	})
}
