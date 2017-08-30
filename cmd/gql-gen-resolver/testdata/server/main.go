package main

import (
	"github.com/lygo/graphql-go"
	"github.com/lygo/graphql-go/cmd/gql-gen-resolver/testdata"
)

type resolver struct {
	testdata.SchemaResolver
}

func main() {
	_, err := graphql.ParseSchema(testdata.Schema, &resolver{})
	if err != nil {
		panic(err)
	}
}
