package elasticsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestFun(t *testing.T) {
	ctx := context.Background()
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Username: "admin",
		Password: "admin",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	assert.NoError(t, err)

	res, err := es.Info()
	if err != nil {
		t.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		t.Fatalf("Error: %s", res.String())
	}

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	fmt.Printf("body: %s", body)

	for i := 0; i < 1000; i++ {
		doc := Fun{
			ID:      gofakeit.UUID(),
			Content: gofakeit.HipsterSentence(10),
		}

		body, err := json.Marshal(&doc)
		assert.NoError(t, err)

		req := esapi.IndexRequest{
			Index:        "fun-index",
			DocumentID:   doc.ID,
			DocumentType: "fun",
			Body:         bytes.NewReader(body),
			Pretty:       true,
		}

		res, err := req.Do(ctx, es)
		assert.NoError(t, err)

		assert.False(t, res.IsError())

		rbody, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)
		fmt.Printf("index response body: %s\n", rbody)
		res.Body.Close()
	}
}
