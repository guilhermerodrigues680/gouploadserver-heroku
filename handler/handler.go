package handler

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/guilhermerodrigues680/gouploadserver/filemanager"
	"github.com/guilhermerodrigues680/gouploadserver/handler/templatetmpl"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type Server struct {
	r                          *httprouter.Router
	logger                     *logrus.Entry
	staticDirPath              string
	keepOriginalUploadFileName bool
}

func NewServer(staticDirPath string, keepOriginalUploadFileName bool, logger *logrus.Entry) *Server {
	router := httprouter.New()
	s := Server{
		r:                          router,
		logger:                     logger,
		staticDirPath:              staticDirPath,
		keepOriginalUploadFileName: keepOriginalUploadFileName,
	}

	router.GET("/*filepath", s.fileHandler)
	router.POST("/*dirpath", s.uploadHandler)

	return &s
}

func (f *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mw := NewLoggingInterceptorOnServer(f.r, f.logger.WithField("server", "interceptor-on-server"))
	mw.ServeHTTP(w, r)
}

func (s *Server) fileHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileUrlPath := p.ByName("filepath")
	filePath := path.Join(s.staticDirPath, ".", fileUrlPath)
	s.logger.Trace(filePath)

	isDir, file, files, err := filemanager.Stat(filePath)
	if err != nil {
		s.logger.Error(err)
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	if isDir {
		// isDir but not HasSuffix '/'
		if !strings.HasSuffix(fileUrlPath, "/") {
			http.Redirect(w, r, fileUrlPath+"/", http.StatusFound)
			return
		}

		// Sort by name
		sort.Slice(files, func(i, j int) bool {
			return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
		})

		t, err := template.New("files").Funcs(template.FuncMap{
			"formatBytes": formatBytes,
		}).Parse(templatetmpl.TemplateListFiles)
		if err != nil {
			s.logger.Errorf("Create template error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		err = t.Execute(w, files)
		if err != nil {
			s.logger.Errorf("Execute template error: %s", err)
			return
		}

		return
	}

	// If is File
	// FIXME comparar desempenho bufio
	buf := make([]byte, 4096) // make a buffer to keep chunks that are read
	ctype, err := getContentType(filePath, buf)
	if err != nil {
		s.logger.Errorf("Get Content-Type error: %s", err)
		return
	}

	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size(), 10))
	s.logger.Tracef("Content-Type: %s, Content-Length: %d", ctype, file.Size())
	w.WriteHeader(http.StatusOK)

	err = filemanager.ReadFileAndWriteToW(w, filePath, buf)
	if err != nil {
		s.logger.Errorf("Read File And Write To W Error: %s", err)
		return
	}
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	dirUrlPath := p.ByName("dirpath")
	dirPath := path.Join(s.staticDirPath, ".", path.Dir(dirUrlPath))
	s.logger.Trace(dirPath)

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		s.logger.Errorf("Parse Media Type error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	s.logger.Trace(mediaType, params)

	boundary := params["boundary"]
	reader := multipart.NewReader(r.Body, boundary)
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				// Done reading body
				break
			}
			s.logger.Errorf("Multipart Reader NextPart error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// only accept fieldname == "file", otherwise, return a validation err
		if part.FormName() != "file" {
			s.logger.Errorf("Field Name != 'file'. Got %s", part.FormName())
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Field Name != 'file'. Got %s", part.FormName())
			return
		}

		contentType := part.Header.Get("Content-Type") // FIXME Not used
		fname := part.FileName()
		s.logger.Infof("multipart/form-data Content-Type: %s, Filename: %s", contentType, fname)

		buf := make([]byte, 4096) // make a buffer to keep chunks that are read
		fileSent, err := filemanager.ReaderToFile(part, dirPath, fname, s.keepOriginalUploadFileName, buf)
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				s.logger.Errorf("Reader To File error, Client closed the connection: %s", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else {
				s.logger.Errorf("Reader To File error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
			return
		}
		s.logger.Infof("File sent: %s", fileSent)
	}
}

// helpers

func getContentType(path string, buf []byte) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		n, _ := io.ReadFull(f, buf) // read a chunk to decide between utf-8 text and binary
		ctype = http.DetectContentType(buf[:n])
		_, err = f.Seek(0, io.SeekStart) // rewind to output whole file
		if err != nil {
			// seeker can't seek
			return "", err
		}
	}

	return ctype, nil
}

func formatBytes(b int64) string {
	const unit = 1024 // ByteCountIEC
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
