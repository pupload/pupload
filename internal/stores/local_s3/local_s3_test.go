package locals3

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// helper to create a store and register cleanup
func newTestStore(t *testing.T) *LocalS3Store {
	t.Helper()

	input := LocalS3StoreInput{BucketName: "test-bucket"}
	store, err := NewLocalS3Store(input)

	if err != nil {
		t.Fatalf("NewLocalS3Store() error = %v", err)
	}

	t.Cleanup(func() {
		store.Close()
	})

	return store
}

func TestNewLocalS3Store_InitializesCorrectly(t *testing.T) {
	store := newTestStore(t)

	if store.server == nil {
		t.Fatalf("expected server to be non-nil")
	}
	if store.client == nil {
		t.Fatalf("expected client to be non-nil")
	}
	if store.bucket != "test-bucket" {
		t.Fatalf("expected bucket %q, got %q", "test-bucket", store.bucket)
	}
}

// round-trip using presigned PUT and GET
func TestLocalS3Store_PutAndGetURL_RoundTrip(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()
	objectName := "folder/test-object.txt"
	expires := 5 * time.Minute
	body := []byte("hello from locals3 test")

	// Get presigned PUT URL
	putURL, err := store.PutURL(ctx, objectName, expires)
	if err != nil {
		t.Fatalf("PutURL() error = %v", err)
	}
	if putURL == nil {
		t.Fatalf("PutURL() = nil URL")
	}
	assertHasSignature(t, putURL)

	// Upload via presigned PUT
	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL.String(), bytes.NewReader(body))
	if err != nil {
		t.Fatalf("creating PUT request: %v", err)
	}
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatalf("PUT to presigned URL failed: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		t.Fatalf("PUT status = %d, want 2xx", putResp.StatusCode)
	}

	// Get presigned GET URL
	getURL, err := store.GetURL(ctx, objectName, expires)
	if err != nil {
		t.Fatalf("GetURL() error = %v", err)
	}
	if getURL == nil {
		t.Fatalf("GetURL() = nil URL")
	}
	assertHasSignature(t, getURL)

	// Download via presigned GET
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, getURL.String(), nil)
	if err != nil {
		t.Fatalf("creating GET request: %v", err)
	}
	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatalf("GET from presigned URL failed: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", getResp.StatusCode, http.StatusOK)
	}

	gotBody, err := io.ReadAll(getResp.Body)
	if err != nil {
		t.Fatalf("reading GET response body: %v", err)
	}

	if !bytes.Equal(gotBody, body) {
		t.Fatalf("downloaded body = %q, want %q", string(gotBody), string(body))
	}
}

func TestLocalS3Store_DeleteObject_MakesObjectInaccessible(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()
	objectName := "to-delete.txt"
	expires := 5 * time.Minute
	body := []byte("delete me")

	// Upload once via presigned PUT
	putURL, err := store.PutURL(ctx, objectName, expires)
	if err != nil {
		t.Fatalf("PutURL() error = %v", err)
	}

	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL.String(), bytes.NewReader(body))
	if err != nil {
		t.Fatalf("creating PUT request: %v", err)
	}
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatalf("PUT to presigned URL failed: %v", err)
	}
	putResp.Body.Close()

	// Confirm we can GET it before delete
	getURLBefore, err := store.GetURL(ctx, objectName, expires)
	if err != nil {
		t.Fatalf("GetURL(before delete) error = %v", err)
	}
	respBefore, err := http.Get(getURLBefore.String())
	if err != nil {
		t.Fatalf("GET before delete failed: %v", err)
	}
	respBefore.Body.Close()
	if respBefore.StatusCode != http.StatusOK {
		t.Fatalf("GET before delete status = %d, want %d", respBefore.StatusCode, http.StatusOK)
	}

	// Delete via API
	if err := store.DeleteObject(ctx, objectName); err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}

	// Now GET should 404 (or at least not 200)
	getURLAfter, err := store.GetURL(ctx, objectName, expires)
	if err != nil {
		t.Fatalf("GetURL(after delete) error = %v", err)
	}
	respAfter, err := http.Get(getURLAfter.String())
	if err != nil {
		// network-level error also acceptable: server closed, etc.
		return
	}
	defer respAfter.Body.Close()

	if respAfter.StatusCode == http.StatusOK {
		t.Fatalf("GET after delete status = %d, want non-200 (typically 404)", respAfter.StatusCode)
	}
}

func TestLocalS3Store_Close_ShutsDownServer(t *testing.T) {
	store := newTestStore(t)

	baseURL := store.server.URL
	store.Close()

	// After Close, requests to the server should fail
	resp, err := http.Get(baseURL + "/")
	if err == nil {
		resp.Body.Close()
		t.Fatalf("expected error when calling closed server, got status %s", resp.Status)
	}
}

// small helper to ensure the URL looks like a signed S3 URL
func assertHasSignature(t *testing.T, u *url.URL) {
	t.Helper()
	q := u.Query()
	if q.Get("X-Amz-Signature") == "" {
		t.Fatalf("presigned URL %q missing X-Amz-Signature", u.String())
	}
}
