package minioutil

import "testing"

func TestPublicURL_basic(t *testing.T) {
	c := &Client{bucket: "mall-media", host: "http://localhost:9000"}
	got := c.PublicURL("products/seed/1.jpg")
	want := "http://localhost:9000/mall-media/products/seed/1.jpg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestNew_trimsTrailingSlash(t *testing.T) {
	c, err := New(Config{
		Endpoint: "localhost:9000", AccessKey: "minioadmin", SecretKey: "minioadmin",
		Bucket: "mall-media", PublicHost: "http://localhost:9000/", UseSSL: false,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.host != "http://localhost:9000" {
		t.Errorf("trailing slash not trimmed: %q", c.host)
	}
}

func TestNew_inferHostFromEndpoint(t *testing.T) {
	c, err := New(Config{
		Endpoint: "minio.local:9000", AccessKey: "x", SecretKey: "y",
		Bucket: "b", UseSSL: false,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.host != "http://minio.local:9000" {
		t.Errorf("inferred host wrong: %q", c.host)
	}
}
