package cca_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	cca "github.com/navigacontentlab/cca-client/v2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func TestClient(t *testing.T) {
	docUUID := os.Getenv("DOC_UUID")
	expectTitle := os.Getenv("EXPECT_TITLE")

	if docUUID == "" {
		t.Skip("no DOC_UUID set")
	}

	ctx := context.Background()
	endpoint, client := testClient(ctx, t)

	docClient := cca.NewDocumentsProtobufClient(endpoint, client)

	res, err := docClient.GetDocument(ctx, &cca.GetDocumentReq{
		UUID: docUUID,
	})
	if err != nil {
		t.Fatalf("failed to get document: %v", err)
	}

	if expectTitle != "" && res.Document.Title != expectTitle {
		t.Errorf("expected document title to be %q, got %q",
			expectTitle, res.Document.Title)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
}

func TestUpload__Filesystem(t *testing.T) {
	ctx := context.Background()
	endpoint, client := testClient(ctx, t)

	files := cca.NewFilesProtobufClient(endpoint, client)

	f, err := os.Open("testdata/sample-image.jpg")
	if err != nil {
		t.Fatalf("failed to open test image: %v", err)
	}

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to stat test image file: %v", err)
	}

	uploadID, res, err := cca.Upload(ctx, files, http.DefaultClient, f, info.Size())
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)

	metadataArt := res.GetArtifact("metadata")
	if metadataArt == nil {
		t.Fatalf("no metadata extracted from upload: %v", err)

		return
	}

	meta, err := files.GetArtifact(ctx, &cca.GetArtifactReq{
		UploadId: uploadID,
		Name:     metadataArt.Name,
	})
	if err != nil {
		t.Fatalf("failed to get extracted metadata: %v", err)
	}

	os.Stderr.Write(meta.Content)
}

func TestUpload__Remote(t *testing.T) {
	ctx := context.Background()
	endpoint, client := testClient(ctx, t)

	files := cca.NewFilesProtobufClient(endpoint, client)

	random, err := http.Get("https://source.unsplash.com/random")
	if err != nil {
		t.Fatalf("failed to fetch random image: %v", err)
	}

	defer func() {
		_ = random.Body.Close()
	}()

	if random.StatusCode != http.StatusOK {
		t.Fatalf("unsplash responded with: %s", random.Status)
	}

	uploadID, res, err := cca.Upload(
		ctx, files, http.DefaultClient,
		random.Body, random.ContentLength,
	)
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	if len(uploadID) == 0 {
		t.Fatalf("invalid upload id: %s", uploadID)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
}

//nolint:thelper
func testClient(ctx context.Context, t *testing.T) (string, *http.Client) {
	t.Helper()

	endpoint := os.Getenv("CCA_ENDPOINT")
	clientID := os.Getenv("CCA_CLIENT_ID")
	clientSecret := os.Getenv("CCA_CLIENT_SECRET")
	tokenEndpoint := os.Getenv("TOKEN_ENDPOINT")

	if clientID == "" {
		t.Skip("no CCA_CLIENT_ID set")
	}

	if clientSecret == "" {
		t.Skip("no CCA_CLIENT_SECRET set")
	}

	if endpoint == "" {
		endpoint = "https://cca-eu-west-1.saas-stage.infomaker.io"
	}

	if tokenEndpoint == "" {
		tokenEndpoint = "https://access-token.stage.imid.infomaker.io/v1/token"
	}

	conf := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenEndpoint,
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	return endpoint, conf.Client(ctx)
}
