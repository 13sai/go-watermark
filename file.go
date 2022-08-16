package gowatermark

import (
	"os"
)

func createDir(dir string) (error) {
	_, err := os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}


func isExistPath(path string) bool {  
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {  
			return true  
		}  
		return false  
	}  
	return true  
}  
