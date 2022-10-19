package util

import "io/ioutil"

func GetFileNames(path string) (filenames []string, err error) {
	ps, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		if !p.IsDir() {
			filenames = append(filenames, p.Name())
		}
	}
	return
}
