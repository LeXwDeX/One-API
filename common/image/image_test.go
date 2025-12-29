package image_test

import (
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/LeXwDeX/one-api/common/client"
	img "github.com/LeXwDeX/one-api/common/image"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "golang.org/x/image/webp"
)

type CountingReader struct {
	reader    io.Reader
	BytesRead int
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.BytesRead += n
	return n, err
}

type imageCase struct {
	name   string
	url    string
	format string
	width  int
	height int
	data   []byte
}

func TestMain(m *testing.M) {
	client.Init()
	m.Run()
}

func setupTestImages(t *testing.T) ([]imageCase, *httptest.Server) {
	t.Helper()
	testFiles := []imageCase{
		{name: "sample.jpeg", format: "jpeg", width: 4, height: 3},
		{name: "sample.png", format: "png", width: 5, height: 4},
		{name: "sample.webp", format: "webp", width: 6, height: 5},
		{name: "sample.gif", format: "gif", width: 7, height: 6},
	}

	mux := http.NewServeMux()
	for i := range testFiles {
		filename := filepath.Join("testdata", testFiles[i].name)
		data, err := os.ReadFile(filename)
		require.NoError(t, err, "failed to read test image %s", filename)
		testFiles[i].data = data

		path := "/" + testFiles[i].name
		contentType := "image/" + testFiles[i].format
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			_, _ = w.Write(data)
		})
	}

	server := httptest.NewServer(mux)
	for i := range testFiles {
		testFiles[i].url = server.URL + "/" + testFiles[i].name
	}
	return testFiles, server
}

func TestDecode(t *testing.T) {
	cases, server := setupTestImages(t)
	defer server.Close()
	// Bytes read: varies sometimes
	// jpeg: 1063892
	// png: 294462
	// webp: 99529
	// gif: 956153
	// jpeg#01: 32805
	for _, c := range cases {
		t.Run("Decode:"+c.format, func(t *testing.T) {
			resp, err := client.UserContentRequestHTTPClient.Get(c.url)
			assert.NoError(t, err)
			defer resp.Body.Close()
			reader := &CountingReader{reader: resp.Body}
			decodedImg, format, err := image.Decode(reader)
			assert.NoError(t, err)
			size := decodedImg.Bounds().Size()
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, size.X)
			assert.Equal(t, c.height, size.Y)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}

	// Bytes read:
	// jpeg: 4096
	// png: 4096
	// webp: 4096
	// gif: 4096
	// jpeg#01: 4096
	for _, c := range cases {
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			resp, err := client.UserContentRequestHTTPClient.Get(c.url)
			assert.NoError(t, err)
			defer resp.Body.Close()
			reader := &CountingReader{reader: resp.Body}
			config, format, err := image.DecodeConfig(reader)
			assert.NoError(t, err)
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, config.Width)
			assert.Equal(t, c.height, config.Height)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}
}

func TestBase64(t *testing.T) {
	cases, server := setupTestImages(t)
	defer server.Close()
	// Bytes read:
	// jpeg: 1063892
	// png: 294462
	// webp: 99072
	// gif: 953856
	// jpeg#01: 32805
	for _, c := range cases {
		t.Run("Decode:"+c.format, func(t *testing.T) {
			data := c.data
			encoded := base64.StdEncoding.EncodeToString(data)
			body := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
			reader := &CountingReader{reader: body}
			decodedImg, format, err := image.Decode(reader)
			assert.NoError(t, err)
			size := decodedImg.Bounds().Size()
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, size.X)
			assert.Equal(t, c.height, size.Y)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}

	// Bytes read:
	// jpeg: 1536
	// png: 768
	// webp: 768
	// gif: 1536
	// jpeg#01: 3840
	for _, c := range cases {
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			data := c.data
			encoded := base64.StdEncoding.EncodeToString(data)
			body := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
			reader := &CountingReader{reader: body}
			config, format, err := image.DecodeConfig(reader)
			assert.NoError(t, err)
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, config.Width)
			assert.Equal(t, c.height, config.Height)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}
}

func TestGetImageSize(t *testing.T) {
	cases, server := setupTestImages(t)
	defer server.Close()
	for i, c := range cases {
		t.Run("Decode:"+strconv.Itoa(i), func(t *testing.T) {
			width, height, err := img.GetImageSize(c.url)
			assert.NoError(t, err)
			assert.Equal(t, c.width, width)
			assert.Equal(t, c.height, height)
		})
	}
}

func TestGetImageSizeFromBase64(t *testing.T) {
	cases, server := setupTestImages(t)
	defer server.Close()
	for i, c := range cases {
		t.Run("Decode:"+strconv.Itoa(i), func(t *testing.T) {
			encoded := base64.StdEncoding.EncodeToString(c.data)
			width, height, err := img.GetImageSizeFromBase64(encoded)
			assert.NoError(t, err)
			assert.Equal(t, c.width, width)
			assert.Equal(t, c.height, height)
		})
	}
}
