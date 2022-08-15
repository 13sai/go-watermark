package gowatermark

import (
	"fmt"
	"os"
)

func createDir(path string) (string, error) {
	var dirs = fmt.Sprintf("%s/", path)
	_, err := os.Stat(dirs)
	if err != nil {
		err = os.MkdirAll(dirs, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return dirs, nil
}
