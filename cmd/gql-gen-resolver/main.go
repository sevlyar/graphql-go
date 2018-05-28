package main

import (
	"bytes"
	"flag"
	"github.com/sevlyar/graphql-go/cmd/gql-gen-resolver/printer"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	schema = flag.String("schema", "", "path to graphql schema")
	out    = flag.String("o", "", "path to out of resolver")
)

func getPackageName(path string) string {
	packageName := filepath.Base(filepath.Dir(path))
	return packageName
}

func openOut() (*os.File, string, error) {
	var path string
	if !filepath.IsAbs(*out) {
		dir, _ := os.Getwd()
		prefix := filepath.SplitList(dir)
		relative := filepath.SplitList(*out)
		prefix = append(prefix, relative...)
		path = filepath.Join(prefix...)
	} else {
		path = *out
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	return file, path, err
}

func main() {
	flag.Parse()

	schemaRaw, err := ioutil.ReadFile(*schema)
	if err != nil {
		log.Fatalf(`can't open schema by path '%s' - err: %s`, *schema, err.Error())
	}

	file, path, err := openOut()
	if err != nil {
		log.Fatalf(`can't open file out by path '%s' - err: %s`, *out, err.Error())
	}
	defer file.Close()

	buff := bytes.NewBuffer(nil)

	packageName := getPackageName(path)
	err = printer.Print(*schema, packageName, string(schemaRaw), buff)
	if err != nil {
		log.Fatalf(`generation failed: %s`, err)
	}

	_, err = file.Write(buff.Bytes())
	if err != nil {
		log.Fatalf(`can't write to file '%s' - err: %s`, *out, err.Error())
	}
	err = file.Sync()
	if err != nil {
		log.Fatalf(`can't write to file '%s' - err: %s`, *out, err.Error())
	}

	cmdOfFmt := exec.Command(`gofmt`, `-w`, path)
	output, err := cmdOfFmt.CombinedOutput()
	if err != nil {
		log.Fatalf("can't call gofmt on file '%s' - err: %s\n%s", output, err.Error(), string(output))
	}
}
