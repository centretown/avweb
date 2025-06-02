package runtime

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func ReadBody(r *http.Request) (id string, key string, val string) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("readBody", err)
	}
	log.Println("Readbody", string(buf))
	lines := strings.Split(string(buf), "&")
	var k, v string
	for _, line := range lines {
		blanksep := strings.Replace(line, "=", " ", 1)
		fmt.Sscan(blanksep, &k, &v)
		if k == "id" {
			id = v
		} else {
			key = k
			val = v
		}
	}
	return
}
