package elasticsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const indexName = "fun-index"

var isIntegration = flag.Bool("i-exec-docker-compose-up", true, "Integration that tests require `docker-compose up`")

func TestFun(t *testing.T) {
	if !*isIntegration {
		t.Skip("Skipping tests because this tests requires `docker-compose up`")
	}

	ctx := context.Background()
	es := mkESClient(t)

	//cleanES(t, es)
	//mkIndex(t, es)

	for i := 0; i < 10; i++ {
		doc := GenFun()
		body, err := json.Marshal(&doc)
		assert.NoError(t, err)

		req := opensearchapi.IndexRequest{
			Index:      indexName,
			DocumentID: doc.ID,
			Body:       bytes.NewReader(body),
			Pretty:     true,
		}

		res, err := req.Do(ctx, es)
		if !assert.NoError(t, err) {
			return
		}

		rbody, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if !assert.NoError(t, err) {
			t.Logf("< %s\n", rbody)
			return
		}
	}
}

func mkESClient(t *testing.T) *opensearch.Client {
	es, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
		Username: "admin",
		Password: "admin",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("mkESClient: %v", err)
	}

	return es
}

func loadFile(t *testing.T, file string) io.Reader {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		t.Fatalf("loadFile: %v", err)
	}

	return f
}

func cleanES(t *testing.T, es *opensearch.Client) {
	ctx := context.Background()
	_, err := opensearchapi.IndicesDeleteRequest{
		Index:          []string{indexName},
		Pretty:         true,
		AllowNoIndices: opensearchapi.BoolPtr(true),
	}.Do(ctx, es)

	if err != nil {
		//if err.Error() != "EOF" {
		t.Fatalf("cleanES: %#v", err)
		//}
	}
}

func mkIndex(t *testing.T, es *opensearch.Client) {
	req := opensearchapi.IndicesCreateRequest{
		Index:  indexName,
		Pretty: true,
		Body:   loadFile(t, "./create-index.json"),
	}

	res, err := req.Do(context.Background(), es)
	assert.NoError(t, err)
	assert.False(t, res.IsError())

	rbody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if !assert.NoError(t, err) {
		t.Logf("mkIndex: %s\n", rbody)
	}
}
