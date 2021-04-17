package transport

import (
	"fmt"
	"gouploadserver"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type Server struct {
	r      *httprouter.Router
	logger *logrus.Entry
}

func NewServer(staticDirPath string, logger *logrus.Entry) *Server {
	router := httprouter.New()

	// router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	//http.ServeFile(w, r, "static/404.html")
	// 	logger.Info("oaoaoao")
	// })

	logReq := NewLoggingInterceptorOnFunc(logger.WithField("server", "interceptor-on-func"))

	// fileServer := http.FileServer(http.Dir(staticDirPath))
	// fs := newFileServer()
	router.GET("/*filepath", logReq.log(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// r.URL.Path = p.ByName("filepath")
		name := "." + p.ByName("filepath")
		logger.Info(name)

		isDir, file, files, err := gouploadserver.Path(name)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(err.Error()))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
			return
		}

		if isDir {
			if !strings.HasSuffix(name, "/") {
				logger.Trace("isDir but not HasSuffix '/'")
				http.Redirect(w, r, name+"/", http.StatusFound)
				return
			}

			logger.Trace("isDir")

			t, err := template.New("files").Parse(templ)
			if err != nil {
				logger.WithError(err).Error()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)

			err = t.Execute(w, files)
			if err != nil {
				logger.WithError(err).Error()
				return
			}

			return
		}

		// Is File
		logger.Trace(file) //FIXME file não usado

		f, err := os.Open(name)
		defer f.Close()

		if err != nil {
			logger.WithError(err).Error()
			return
		}

		ctype := mime.TypeByExtension(filepath.Ext(name))
		if ctype == "" {
			// read a chunk to decide between utf-8 text and binary
			var buf [512]byte
			n, _ := io.ReadFull(f, buf[:])
			ctype = http.DetectContentType(buf[:n])
			_, err = f.Seek(0, io.SeekStart) // rewind to output whole file
			if err != nil {
				http.Error(w, "seeker can't seek", http.StatusInternalServerError)
				return
			}
		}
		logger.Trace("Content-Type", ctype)
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(http.StatusOK)

		// make a buffer to keep chunks that are read
		buf := make([]byte, 1024)
		// var buf [1024]byte
		for {
			// read a chunk
			n, err := f.Read(buf)
			if err != nil && err != io.EOF {
				logger.WithError(err).Error()
				return
			}
			if n == 0 {
				break
			}

			// write a chunk
			if _, err := w.Write(buf[:n]); err != nil {
				logger.WithError(err).Error()
				return
			}
		}

		// fmt.Fprint(w, isDir, files, file)

		// fileServer.ServeHTTP(w, req)
		// fs.ServeHTTP(w, r)
	}))

	router.POST("/upload", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		// r.ParseMultipartForm(10 << 20)
		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			logger.WithError(err).Error()
			return
		}
		logger.Trace(mediaType, params)

		boundary := params["boundary"]
		reader := multipart.NewReader(r.Body, boundary)
		for {
			part, err := reader.NextPart()
			if err != nil {
				if err == io.EOF {
					// Done reading body
					break
				}
				logger.WithError(err).Error()
				return
			}

			if part.FormName() != "file" {
				// return a validation err
				logger.Error("FormNane != file")
				return
			}

			contentType := part.Header.Get("Content-Type")
			fname := part.FileName()
			// part is an io.Reader, deal with it
			logger.Trace(contentType, fname)

			// Create a temporary file within our tmp--upload directory that follows
			// a particular naming pattern
			tempFile, err := ioutil.TempFile(".", "*-"+fname)
			if err != nil {
				fmt.Println(err)
			}
			defer tempFile.Close()

			startTime := time.Now()

			buf := make([]byte, 1024)
			for {
				// read a chunk
				n, err := part.Read(buf)
				if err != nil && err != io.EOF {
					logger.WithError(err).Error()
					return
				}

				if n == 0 {
					break
				}

				// write a chunk
				if _, err := tempFile.Write(buf[:n]); err != nil {
					logger.WithError(err).Error()
					return
				}
			}

			logger.Infof("Total Upload Time: %s", time.Since(startTime))
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
	})

	return &Server{
		r:      router,
		logger: logger,
	}
}

func (f *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mw := NewLoggingInterceptorOnServer(f.r, f.logger.WithField("server", "interceptor-on-server"))
	mw.ServeHTTP(w, r)
}

const templ = `<!DOCTYPE html>
<html>
  <head>
    <title>File list</title>
  </head>
  <body>
  	<h1>Upload de arquivo</h1>
    <form action="/upload" method="post" enctype="multipart/form-data">
        <label>
            Selecione o Arquivo
            <input type="file" name="file">
        </label>
        <br>
        <input type="submit" value="Enviar">
    </form>
	<br>
    <p>
      Files
    </p>
    <table>
      <tr>
        <td>File</td>
        <td>Size</td>
    	</tr>
      {{ range . }}
        <tr>
		<td>
			{{if .IsDir}}
				<a href="{{ .Name }}/">{{ .Name }}/</a>
			{{ else }}
				<a href="{{ .Name }}">{{ .Name }}</a>
			{{ end }}
		</td>
          <td>{{ .Size }}</td>
        </tr>
      {{ end }} 
    </table>
  </body>
</html>`
