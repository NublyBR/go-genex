package genex

import (
	"bytes"
	"regexp"
	"testing"
)

func TestCompiler(t *testing.T) {
	_, err := Compile("<64>{1}[a-z]{2}(a{3}b|c(d)){4}")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGeneration(t *testing.T) {
	var tests = []string{
		`[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}`,
		`[\x21-\x7e]{64}`,
		`((\d|([1-9]\d)|(1\d{2})|(2[0-4]\d)|(25[0-5]))\.){3}(\d|([1-9]\d)|(1\d{2})|(2[0-4]\d)|(25[0-5]))`,
		`([0-9a-f]{4}:){7}[0-9a-f]{4}`,
		`flag\{[0-9a-f]{32}\}`,
		`(\d{3}\.){2}\d{3}-\d{3}`,
		`\d{2}(\.\d{3}){2}/0001-\d{2}`,
		`([0-9A-Z]{5}-){4}[0-9A-Z]{5}`,
		`[a-bd-eg-hj-k]{64}`,
		`(abc|def|ghi|\d{3}|[j-z]{3}){9}`,
		`a?b?c?d?e?f?g?h?i?`,
		`\w{16}\W{16}\d{16}`,
		`\s{16}\S{16}`,
	}

	var reNormalize = regexp.MustCompile(`\033\[[^m]+m`)

	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	for _, test := range tests {
		gen, err := Compile(test)
		if err != nil {
			t.Fatal(err)
		}

		mt, err := regexp.Compile(`(?m)^` + test + `$`)
		if err != nil {
			t.Fatal(err)
		}

		for range 100 {
			gen.Sample(buf)
			if !mt.Match(buf.Bytes()) {
				t.Errorf("sampled text %q failed to match regex %q", buf.String(), mt.String())
			}
			buf.Reset()
		}

		// Test if the string representation of the generator can be used to recompile the same generator.

		cmp := reNormalize.ReplaceAllLiteralString(gen.String(), "")
		recmp, err := Compile(cmp)
		if err != nil {
			t.Fatalf("expected %q repr %q to be recompilable, got error %v", test, cmp, err)
		}

		for range 100 {
			recmp.Sample(buf)
			if !mt.Match(buf.Bytes()) {
				t.Errorf("sampled text %q failed to match regex %q", buf.String(), mt.String())
			}
			buf.Reset()
		}
	}
}
