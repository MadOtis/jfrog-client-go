package utils

import (
	"reflect"
	"sort"
	"testing"

	"github.com/madotis/jfrog-client-go/utils/io"
	"github.com/stretchr/testify/assert"
)

func TestRemoveRepoFromPath(t *testing.T) {
	assertRemoveRepoFromPath("repo/abc/def", "/abc/def", t)
	assertRemoveRepoFromPath("repo/(*)", "/(*)", t)
	assertRemoveRepoFromPath("repo/", "/", t)
	assertRemoveRepoFromPath("/abc/def", "/abc/def", t)
	assertRemoveRepoFromPath("aaa", "aaa", t)
	assertRemoveRepoFromPath("", "", t)
}

func assertRemoveRepoFromPath(path, expected string, t *testing.T) {
	result := removeRepoFromPath(path)
	if expected != result {
		t.Error("Unexpected string built by removeRepoFromPath. Expected: `" + expected + "` Got `" + result + "`")
	}
}

func TestBuildTargetPath(t *testing.T) {
	assertBuildTargetPath("1(*)234", "1hello234", "{1}", "hello", true, t)
	assertBuildTargetPath("1234", "1hello234", "{1}", "{1}", true, t)
	assertBuildTargetPath("1(2*5)6", "123456", "{1}", "2345", true, t)
	assertBuildTargetPath("(*) something", "doing something", "{1} something else", "doing something else", true, t)
	assertBuildTargetPath("(switch) (this)", "switch this", "{2} {1}", "this switch", true, t)
	assertBuildTargetPath("before(*)middle(*)after", "before123middle456after", "{2}{1}{2}", "456123456", true, t)
	assertBuildTargetPath("foo/before(*)middle(*)after", "foo/before123middle456after", "{2}{1}{2}", "456123456", true, t)
	assertBuildTargetPath("foo/before(*)middle(*)after", "bar/before123middle456after", "{2}{1}{2}", "456123456", true, t)
	assertBuildTargetPath("foo/before(*)middle(*)after", "bar/before123middle456after", "{2}{1}{2}", "{2}{1}{2}", false, t)
	assertBuildTargetPath("foo/before(*)middle(*)", "bar/before123middle456after", "{2}{1}{2}", "456after123456after", true, t)
	assertBuildTargetPath("f(*)oo/before(*)after", "f123oo/before456after", "{2}{1}{2}", "456123456", true, t)
	assertBuildTargetPath("f(*)oo/before(*)after", "f123oo/before456after", "{2}{1}{2}", "456123456", false, t)
	assertBuildTargetPath("generic-(*)-(bar)", "generic-foo-bar/after/a.in", "{1}/{2}", "foo/bar", true, t)
	assertBuildTargetPath("generic-(*)-(bar)/(*)", "generic-foo-bar/after/a.in", "{1}/{2}/{3}", "foo/bar/after/a.in", true, t)
	assertBuildTargetPath("generic-(*)-(bar)", "generic-foo-bar/after/a.in", "{1}/{2}/after/a.in", "foo/bar/after/a.in", true, t)
	assertBuildTargetPath("", "nothing should change", "nothing should change", "nothing should change", true, t)
}

func assertBuildTargetPath(regexp, source, dest, expected string, ignoreRepo bool, t *testing.T) {
	result, _, err := BuildTargetPath(regexp, source, dest, ignoreRepo)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSplitWithEscape(t *testing.T) {
	assertSplitWithEscape("", []string{""}, t)
	assertSplitWithEscape("a", []string{"a"}, t)
	assertSplitWithEscape("a/b", []string{"a", "b"}, t)
	assertSplitWithEscape("a/b/c", []string{"a", "b", "c"}, t)
	assertSplitWithEscape("a/b\\5/c", []string{"a", "b5", "c"}, t)
	assertSplitWithEscape("a/b\\\\5.2/c", []string{"a", "b\\5.2", "c"}, t)
	assertSplitWithEscape("a\\8/b\\5/c", []string{"a8", "b5", "c"}, t)
	assertSplitWithEscape("a\\\\8/b\\\\5.2/c", []string{"a\\8", "b\\5.2", "c"}, t)
	assertSplitWithEscape("a/b\\5/c\\0", []string{"a", "b5", "c0"}, t)
	assertSplitWithEscape("a/b\\\\5.2/c\\\\0", []string{"a", "b\\5.2", "c\\0"}, t)
}

func assertSplitWithEscape(str string, expected []string, t *testing.T) {
	result := SplitWithEscape(str, '/')
	if !reflect.DeepEqual(result, expected) {
		t.Error("Unexpected string array built. Expected: `", expected, "` Got `", result, "`")
	}
}

func TestConvertLocalPatternToRegexp(t *testing.T) {
	var tests = []struct {
		localPath string
		expected  string
	}{
		{"./", "^.*$"},
		{".\\\\", "^.*$"},
		{".\\", "^.*$"},
		{"./abc", "abc"},
		{".\\\\abc", "abc"},
		{".\\abc", "abc"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, ConvertLocalPatternToRegexp(test.localPath, RegExp))
	}
}
func TestCleanPath(t *testing.T) {
	if io.IsWindows() {
		parameter := "\\\\foo\\\\baz\\\\..\\\\bar\\\\*"
		got := cleanPath(parameter)
		want := "\\\\foo\\\\bar\\\\*"
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "\\\\foo\\\\\\\\bar\\\\*"
		got = cleanPath(parameter)
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "\\\\foo\\\\.\\\\bar\\\\*"
		got = cleanPath(parameter)
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "\\\\foo\\\\.\\\\bar\\\\*\\\\"
		want = "\\\\foo\\\\bar\\\\*\\\\"
		got = cleanPath(parameter)
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "foo\\\\bar"
		got = cleanPath(parameter)
		want = "foo\\\\bar"
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = ".\\\\foo\\\\bar\\\\"
		got = cleanPath(parameter)
		want = "foo\\\\bar\\\\"
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
	} else {
		parameter := "/foo/bar/"
		got := cleanPath(parameter)
		want := "/foo/bar/"
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "/foo/baz/../bar/*"
		got = cleanPath(parameter)
		want = "/foo/bar/*"
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "/foo//bar/*"
		got = cleanPath(parameter)
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "/foo/./bar/*"
		got = cleanPath(parameter)
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "/foo/./bar/*/"
		want = "/foo/bar/*/"
		got = cleanPath(parameter)
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "foo/bar"
		got = cleanPath(parameter)
		want = "foo/bar"
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
		parameter = "./foo/bar/"
		got = cleanPath(parameter)
		want = "foo/bar/"
		if got != want {
			t.Errorf("cleanPath(%s) == %s, want %s", parameter, got, want)
		}
	}
}
func TestIsWildcardParentheses(t *testing.T) {
	strA := "/tmp/cache/download/(github.com/)"
	strB := "/tmp/cache/download/(github.com/*)"
	parenthesesA := CreateParenthesesSlice(strA, "")
	parenthesesB := CreateParenthesesSlice(strA, "{1}")

	got := isWildcardParentheses(strA, parenthesesA)
	want := false
	if got != want {
		t.Errorf("TestIsWildcardParentheses() == %t, want %t", got, want)
	}

	got = isWildcardParentheses(strB, parenthesesB)
	want = true
	if got != want {
		t.Errorf("TestIsWildcardParentheses() == %t, want %t", got, want)
	}
}

func equalSlicesIgnoreOrder(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	sort.Strings(s1)
	sort.Strings(s2)
	return reflect.DeepEqual(s1, s2)
}

func TestGetMaxPlaceholderIndex(t *testing.T) {
	type args struct {
		toReplace string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{"empty", args{""}, 0, nil},
		{"empty", args{"{}"}, 0, nil},
		{"basic", args{"{1}{5}{3}"}, 5, nil},
		{"basic", args{"}5{{3}"}, 3, nil},
		{"basic", args{"{1}}5}{3}"}, 3, nil},
		{"basic", args{"{1}5{}}{3}"}, 3, nil},
		{"special characters", args{"!@#$%^&*abc(){}}{{2}!@#$%^&*abc(){}}{{1}!@#$%^&*abc(){}}{"}, 2, nil},
		{"multiple digits", args{"{2}{100}fdsff{101}d#%{99}"}, 101, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMaxPlaceholderIndex(tt.args.toReplace)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "getMaxPlaceholderIndex(%v)", tt.args.toReplace)
		})
	}
}

func TestReplacePlaceHolders(t *testing.T) {
	type args struct {
		groups    []string
		toReplace string
		isRegexp  bool
	}
	tests := []struct {
		name            string
		args            args
		expected        string
		expectedBoolean bool
	}{
		// First element in the group isn't relevant cause the matching loop starts from index 1.
		{"non regexp, empty group", args{[]string{}, "{1}-{2}-{3}", false}, "{1}-{2}-{3}", false},
		{"non regexp, empty group", args{[]string{""}, "{1}-{2}-{3}", false}, "{1}-{2}-{3}", false},
		{"regexp, empty group", args{[]string{}, "{1}-{2}-{3}", true}, "{1}-{2}-{3}", false},
		{"regexp, empty group", args{[]string{""}, "{1}-{2}-{3}", true}, "{1}-{2}-{3}", false},
		// Non regular expressions
		{"basic", args{[]string{"", "a", "b", "c"}, "{1}-{2}-{3}", false}, "a-b-c", true},
		{"opposite order", args{[]string{"", "a", "b", "c"}, "{3}-{2}-{1}-{4}", false}, "c-b-a-{4}", true},
		{"double", args{[]string{"", "a", "b"}, "{2}-{2}-{1}-{1}", false}, "b-b-a-a", true},
		{"skip placeholders indexes", args{[]string{"", "a", "b"}, "{4}-{1}", false}, "b-a", true},
		// Regular expressions
		{"basic", args{[]string{"", "a", "b", "c"}, "{1}-{2}-{3}", true}, "a-b-c", true},
		{"opposite order", args{[]string{"", "a", "b", "c"}, "{4}-{3}-{2}-{5}", true}, "{4}-c-b-{5}", true},
		{"double", args{[]string{"", "a", "b"}, "{2}-{2}-{1}-{1}", true}, "b-b-a-a", true},
		{"skip placeholders indexes", args{[]string{"", "a", "b"}, "{3}-{1}", true}, "{3}-a", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, replaceOccurred, err := ReplacePlaceHolders(tt.args.groups, tt.args.toReplace, tt.args.isRegexp)
			assert.NoError(t, err)
			assert.Equalf(t, tt.expected, result, "ReplacePlaceHolders(%v, %v, %v)", tt.args.groups, tt.args.toReplace, tt.args.isRegexp)
			assert.Equalf(t, tt.expectedBoolean, replaceOccurred, "ReplacePlaceHolders(%v, %v, %v)", tt.args.groups, tt.args.toReplace, tt.args.isRegexp)
		})
	}
}
