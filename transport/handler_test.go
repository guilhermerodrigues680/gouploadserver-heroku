package transport

import (
	"bytes"
	"gouploadserver"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestFileHandlerStatusFound(t *testing.T) {
	// https://blog.questionable.services/article/testing-http-handlers-go/
	req, err := http.NewRequest(http.MethodGet, "/cmd", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s := NewServer("..", logrus.WithField("test", true))

	s.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusFound {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if location := rr.HeaderMap.Get("Location"); location != "/cmd/" {
		t.Fatalf("handler returned wrong header Location: got %v want %v", location, "/cmd/")
	}
}

func TestFileHandlerStatusOK(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/cmd/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s := NewServer("..", logrus.WithField("test", true))

	s.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestFileHandlerStream(t *testing.T) {
	filepath := "/mime-type-test/yolinux-mime-test.gif"
	req, err := http.NewRequest(http.MethodGet, filepath, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s := NewServer("..", logrus.WithField("test", true))

	s.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "image/gif" {
		t.Fatalf("handler returned wrong header Location: got %v want %v", contentType, "image/gif")
	}

	var bufferR bytes.Buffer
	name := path.Join(s.staticDirPath, ".") + filepath
	buf := make([]byte, 4096)
	gouploadserver.ReadFileAndWriteToW(&bufferR, name, buf)

	bufferF, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(bufferR.Bytes(), bufferF) {
		t.Fatalf("handler returned wrong body: got file size %v want file size %v", bufferR.Len(), len(bufferF))
	}
}

func TestUploadHandlerStream(t *testing.T) {
	s := NewServer("..", logrus.WithField("test", true))
	filepath := "/mime-type-test/yolinux-mime-test.gif"

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	f, err := os.Open(path.Join(s.staticDirPath, ".") + filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var fw io.Writer
	fw, err = w.CreateFormFile("file", path.Base(f.Name()))
	if err != nil {
		t.Fatal(err)
	}

	if _, err = io.Copy(fw, f); err != nil {
		t.Fatal(err)
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	req, err := http.NewRequest(http.MethodPost, "/upload", &b)
	if err != nil {
		t.Fatal(err)
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func BenchmarkUploadHandlerStream(b *testing.B) {
	s := NewServer("..", logrus.WithField("test", true))
	filepath := "/mime-type-test/yolinux-mime-test.gif"

	for n := 0; n < b.N; n++ {
		var buffer bytes.Buffer
		w := multipart.NewWriter(&buffer)
		f, err := os.Open(path.Join(s.staticDirPath, ".") + filepath)
		if err != nil {
			b.Fatal(err)
		}
		defer f.Close()

		var fw io.Writer
		fw, err = w.CreateFormFile("file", path.Base(f.Name()))
		if err != nil {
			b.Fatal(err)
		}

		if _, err = io.Copy(fw, f); err != nil {
			b.Fatal(err)
		}

		// Don't forget to close the multipart writer.
		// If you don't close it, your request will be missing the terminating boundary.
		w.Close()

		req, err := http.NewRequest(http.MethodPost, "/upload", &buffer)
		if err != nil {
			b.Fatal(err)
		}
		// Don't forget to set the content type, this will contain the boundary.
		req.Header.Set("Content-Type", w.FormDataContentType())

		rr := httptest.NewRecorder()
		s.ServeHTTP(rr, req)
	}

	// if status := rr.Code; status != http.StatusOK {
	// 	b.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	// }
}

func BenchmarkFileHandlerStream(b *testing.B) {
	filepath := "/mime-type-test/large-file.mp4"
	req, err := http.NewRequest(http.MethodGet, filepath, nil)
	if err != nil {
		b.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s := NewServer("..", logrus.WithField("test", true))

	for n := 0; n < b.N; n++ {
		s.ServeHTTP(rr, req)
	}
}
