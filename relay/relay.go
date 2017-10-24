package relay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	graphql "github.com/lygo/graphql-go"
)

func MarshalID(kind string, spec interface{}) graphql.ID {
	d, err := json.Marshal(spec)
	if err != nil {
		panic(fmt.Errorf("relay.MarshalID: %s", err))
	}
	return graphql.ID(base64.URLEncoding.EncodeToString(append([]byte(kind+":"), d...)))
}

func UnmarshalKind(id graphql.ID) string {
	s, err := base64.URLEncoding.DecodeString(string(id))
	if err != nil {
		return ""
	}
	i := strings.IndexByte(string(s), ':')
	if i == -1 {
		return ""
	}
	return string(s[:i])
}

func UnmarshalSpec(id graphql.ID, v interface{}) error {
	s, err := base64.URLEncoding.DecodeString(string(id))
	if err != nil {
		return err
	}
	i := strings.IndexByte(string(s), ':')
	if i == -1 {
		return errors.New("invalid graphql.ID")
	}
	return json.Unmarshal([]byte(s[i+1:]), v)
}

type Handler struct {
	Schema *graphql.Schema
}

type Params struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params Params
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// set parameters to context
	ctx := WithParams(r.Context(), params)

	response := h.Schema.Exec(ctx, params.Query, params.OperationName, params.Variables)

	// check additional queries from context
	if queries, ok := GetQueries(ctx); ok {
		for topic, params := range queries {
			addResponse := h.Schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
			// TODO: don't ignore error
			if response.Extensions == nil {
				response.Extensions = make(map[string]interface{})
			}
			response.Extensions[topic] = addResponse.Data
		}
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

type ctxKey int

const (
	paramsKey ctxKey = iota
	queriesKey
)

func WithParams(ctx context.Context, params Params) context.Context {
	return context.WithValue(ctx, paramsKey, params)
}

func GetParams(ctx context.Context) (Params, bool) {
	v, ok := ctx.Value(paramsKey).(Params)
	return v, ok
}

func WithQueries(ctx context.Context, queries map[string]Params) context.Context {
	return context.WithValue(ctx, queriesKey, queries)
}

func GetQueries(ctx context.Context) (map[string]Params, bool) {
	v, ok := ctx.Value(queriesKey).(map[string]Params)
	return v, ok
}
