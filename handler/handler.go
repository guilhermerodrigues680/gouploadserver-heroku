package handler

import (
	"fmt"
	"gouploadserver/filemanager"
	"gouploadserver/handler/templatetmpl"
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
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type Server struct {
	r             *httprouter.Router
	logger        *logrus.Entry
	staticDirPath string
}

func NewServer(staticDirPath string, logger *logrus.Entry) *Server {
	router := httprouter.New()
	s := Server{
		r:             router,
		logger:        logger,
		staticDirPath: staticDirPath,
	}

	// router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	//http.ServeFile(w, r, "static/404.html")
	// 	logger.Info("oaoaoao")
	// })

	// FIXME remover este log
	logReq := NewLoggingInterceptorOnFunc(logger.WithField("server", "interceptor-on-func"))

	// fileServer := http.FileServer(http.Dir(staticDirPath))
	// fs := newFileServer()
	router.GET("/*filepath", logReq.log(s.fileHandler))
	router.POST("/*uploadpath", s.uploadHandler)

	return &s
}

func (f *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mw := NewLoggingInterceptorOnServer(f.r, f.logger.WithField("server", "interceptor-on-server"))
	mw.ServeHTTP(w, r)
}

func (s *Server) fileHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// r.URL.Path = p.ByName("filepath")
	name := path.Join(s.staticDirPath, ".") + p.ByName("filepath")
	s.logger.Info(name)

	isDir, file, files, err := filemanager.Path(name)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	if isDir {
		if !strings.HasSuffix(name, "/") {
			s.logger.Trace("isDir but not HasSuffix '/'")
			http.Redirect(w, r, name+"/", http.StatusFound)
			return
		}

		sort.Slice(files, func(i, j int) bool {
			return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
		})

		s.logger.Trace("isDir")

		t, err := template.New("files").Funcs(template.FuncMap{
			"formatBytes": formatBytes,
		}).Parse(templatetmpl.TemplateListFiles)
		if err != nil {
			s.logger.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		err = t.Execute(w, files)
		if err != nil {
			s.logger.WithError(err).Error()
			return
		}

		return
	}

	// Is File
	s.logger.Trace(file) //FIXME file não usado

	// 1 - Saber Mime-Type
	// 2 - Enviar arquivo

	// FIXME comparar desempenho bufio
	buf := make([]byte, 4096) // make a buffer to keep chunks that are read
	ctype, err := getContentType(name, buf)
	if err != nil {
		s.logger.WithError(err).Error()
		return
	}

	s.logger.Trace("Content-Type", ctype)
	s.logger.Trace("Content-Type", ctype)
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size(), 10))
	w.WriteHeader(http.StatusOK)

	filemanager.ReadFileAndWriteToW(w, name, buf)
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	// r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file

	// r.URL.Path = p.ByName("filepath")
	uploadPath := path.Join(s.staticDirPath, ".", path.Dir(p.ByName("uploadpath")))
	s.logger.Info(uploadPath)

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		s.logger.WithError(err).Error()
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
			s.logger.WithError(err).Error()
			return
		}

		if part.FormName() != "file" {
			// return a validation err
			s.logger.Errorf("FormName != file. got %v want %v", part.FormName(), "file")
			return
		}

		// FIXME Nao usado
		contentType := part.Header.Get("Content-Type")
		fname := part.FileName()
		s.logger.Trace(contentType, fname)

		startTime := time.Now()
		buf := make([]byte, 4096) // make a buffer to keep chunks that are read
		filemanager.ReaderToFile(part, uploadPath, fname, buf)
		s.logger.Infof("Total Upload Time: %s", time.Since(startTime))
	}

	// return

	// reqFile, handler, err := r.FormFile("file")
	// if err != nil {
	// 	fmt.Println("Error Retrieving the File")
	// 	fmt.Println(err)
	// 	return
	// }
	// defer reqFile.Close()
	// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	// fmt.Printf("File Size: %+v\n", handler.Size)
	// fmt.Printf("MIME Header: %+v\n", handler.Header)

	// return that we have successfully uploaded our file!
	//fmt.Fprint(w, "Successfully Uploaded File")
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
