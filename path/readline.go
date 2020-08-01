package utils

import (
	"bufio"
	"os"
)

func ReadLine(path string, fn func(line string)) error {

	if fn == nil {
		return nil
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fn(scanner.Text())
	}

	err = scanner.Err()
	return err
}
