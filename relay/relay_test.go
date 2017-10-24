package relay_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lygo/graphql-go"
	"github.com/lygo/graphql-go/example/starwars"
	"github.com/lygo/graphql-go/relay"
)

func TestServeHTTP(t *testing.T) {
	starwarsSchema := graphql.MustParseSchema(starwars.Schema, &starwars.Resolver{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/some/path/here", strings.NewReader(`{"query":"{ hero { name } }", "operationName":"", "variables": null}`))
	h := relay.Handler{Schema: starwarsSchema}

	ctx := relay.WithQueries(r.Context(), map[string]relay.Params{
		"addition": {
			Query: `{ hero(episode: EMPIRE) { name } }`,
		},
	})

	h.ServeHTTP(w, r.WithContext(ctx))

	if w.Code != 200 {
		t.Fatalf("Expected status code 200, got %d.", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("Invalid content-type. Expected [application/json], but instead got [%s]", contentType)
	}

	expectedResponse := `{"data":{"hero":{"name":"R2-D2"}},"extensions":{"addition":{"hero":{"name":"Luke Skywalker"}}}}`
	actualResponse := w.Body.String()
	if expectedResponse != actualResponse {
		t.Fatalf("Invalid response. Expected [%s], but instead got [%s]", expectedResponse, actualResponse)
	}
}
