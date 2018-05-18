package filekit

import (
	"bufio"
	"fmt"
	"os"
)

func WritePropertiesFile(path string, data map[string]interface{}) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	w := bufio.NewWriter(f)
	for k, v := range data {
		w.WriteString(fmt.Sprintf("%s=%v\n", k, v))
	}
	err = w.Flush()
	return
}
