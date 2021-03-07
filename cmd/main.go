package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func initZil() {
	os.Mkdir(".zil", 0755)
	os.Mkdir(".zil/objects", 0755)
	os.Mkdir(".zil/tree", 0755)
}
func Err(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
func makeObj(t int, magicNumber int, content []byte) {
	sha := sha1.New()
	sha.Write([]byte{byte(magicNumber), 00})
	sha.Write(content)

	hex := hex.EncodeToString(sha.Sum([]byte{}))
	err := os.Mkdir(".zil/objects/"+hex[:2], 0755)
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write([]byte{byte(t)})
	gz.Write(content)
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	os.WriteFile(".zil/objects/"+hex[:2]+"/"+hex[3:], b.Bytes(), 0060)
	Err(err)

}

type obj struct {
	magicNumber int
	name        string
	content     string
}

func readObj(sha string) obj {
	path := ".zil/objects/" + sha[:2] + "/" + sha[2:]
	os.Chmod(path, 0755)
	defer os.Chmod(path, 0060)
	body, fErr := os.ReadFile(path)
	Err(fErr)
	gz, _ := gzip.NewReader(bytes.NewReader(body))
	d, bErr := ioutil.ReadAll(gz)
	Err(bErr)
	headerNum := 0
	fName := ""
	content := ""
	for i, b := range d {
		if i == 0 {
			headerNum = int(b)
			continue
		}
		if i <= headerNum+1 {
			fName += string(b)
			continue
		}
		content += string(b)
	}
	return obj{magicNumber: 0, name: fName, content: content}
}
func writeStage() {

}
func main() {
	initZil()
	content := []byte{}
	content = append(content, []byte("hi.js")...)
	content = append(content, 00)
	content = append(content, []byte("HELLO woRLD!!")...)
	//makeObj(4, content)
	ob := readObj("beaa9211a631fe91f8765c1ca4b29048f885ef9")
	fmt.Println(ob)
}
