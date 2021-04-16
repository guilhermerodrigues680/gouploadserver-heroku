package gouploadserver

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
)

var ErrNotFound error = errors.New("Not Found")

func Path(path string) (bool, fs.FileInfo, []fs.FileInfo, error) {

	if path == "" {
		path = "."
	}

	// O path existe?
	fi, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Arquivo/Diretorio nao existe", err)
		} else {
			fmt.Println("Erro nao reconhecido: ", err)
		}
		return false, nil, nil, err
	}
	fmt.Print(fi)

	// É um diretório?
	// if fi.Mode().IsDir() && !strings.HasSuffix(path, "/") {
	// }

	switch mode := fi.Mode(); {
	case mode.IsDir():
		// do directory stuff
		fmt.Println("directory")
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return true, nil, nil, err
		}
		return true, nil, files, nil
	case mode.IsRegular():
		// do file stuff
		fmt.Println("file")
		// file, err := ioutil.ReadFile(path)
		// if err != nil {
		// 	return true, nil, nil, err
		// }
		// return true, nil, file, nil
		return false, fi, nil, nil
	}

	// files, err := ioutil.ReadDir(".")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // file := files[0]
	// for _, file := range files {
	// 	// w.Write([]byte(file.Name() + "\n"))
	// 	fmt.Println(file.Name())
	// 	file.Size()
	// }

	return false, nil, nil, nil
}
