package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os/exec"
	"testing"
)

func TestMainFunction(t *testing.T) {
	var err error
	err = flag.Set("schema", "testdata/schema.gql")
	if err != nil {
		t.Fatal(err)
	}
	err = flag.Set("o", "testdata/schema.gql.go")
	if err != nil {
		t.Fatal(err)
	}

	main()

	got, err := ioutil.ReadFile(`testdata/schema.gql.go`)
	if err != nil {
		t.Fatal(err)
	}

	exp, err := ioutil.ReadFile(`testdata/expected.gql.schema`)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(exp, got) {
		cmd := exec.Command(
			`diff`,
			`-y`,
			`testdata/expected.gql.schema`,
			`testdata/schema.gql.go`,
		)
		info, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err, string(info))
		}
		t.Error("expected the same resolver, but thay differens:\n\n" + string(info))
	}

	cmd := exec.Command(
		`go`,
		`run`,
		`./testdata/server/main.go`,
	)
	info, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(info))
	}
}
