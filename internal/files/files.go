package files

import "os"

func Write(f, s string) error {
	file, err := os.OpenFile(f, os.O_CREATE|os.O_WRONLY, 0o660)
	if err != nil {
		return err
	}

	file.WriteString(s)
	file.Close()
	return nil
}

func Read(f string) (string, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
