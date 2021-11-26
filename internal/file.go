package internal

import (
	"io/fs"
	"io/ioutil"
)

type FileRepository struct {}

func (f *FileRepository) Write(filename string, data []byte, perm fs.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}