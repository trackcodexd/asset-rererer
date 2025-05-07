package files

import (
	"os"
	"path/filepath"
)

func Write(n, c string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(wd, n), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(c)
	return err
}

func Read(n string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filepath.Join(wd, n))
	return string(data), err
}
