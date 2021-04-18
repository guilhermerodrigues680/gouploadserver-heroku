package filemanager

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
)

var ErrUnrecognizedMode error = errors.New("Unrecognized mode")

func Stat(path string) (bool, fs.FileInfo, []fs.FileInfo, error) {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return false, nil, nil, err
	}

	switch mode := fileinfo.Mode(); {
	case mode.IsDir():
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return true, nil, nil, err
		}
		return true, nil, files, nil
	case mode.IsRegular():
		return false, fileinfo, nil, nil
	default:
		return false, nil, nil, fmt.Errorf("%w %s", ErrUnrecognizedMode, mode)
	}
}

func ReadFileAndWriteToW(w io.Writer, path string, buf []byte) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		// read a chunk
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

func ReaderToFile(r io.Reader, dir string, fname string, keepOriginalFileName bool, buf []byte) (string, error) {
	// FIXME file permissions originais

	ext := path.Ext(fname)
	name := fname[0 : len(fname)-len(ext)]
	tempFilePattern := name + "-*" + ext
	tempFile, err := ioutil.TempFile(dir, tempFilePattern)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			os.Remove(tempFile.Name())
			return "", err
		}

		if n == 0 {
			break
		}

		// write a chunk
		if _, err := tempFile.Write(buf[:n]); err != nil {
			os.Remove(tempFile.Name())
			return "", err
		}
	}

	if keepOriginalFileName {
		finalFileName := path.Join(dir, fname)
		err = os.Rename(tempFile.Name(), path.Join(dir, fname))
		if err != nil {
			return "", err
		}
		return finalFileName, nil
	}

	return tempFile.Name(), nil
}
