package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	//"github.com/waigani/diffparser"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	//zi "github.com/ashtyn3/zi/pkg"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Initializes a zil repo
func initZil() {
	os.Mkdir(".zil", 0755)
	os.Mkdir(".zil/objects", 0755)
	os.Mkdir(".zil/tree", 0755)
	os.WriteFile(".zil/ROOF", []byte{}, 0060)
}
func Err(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
func assembleObjHeader(s int, name, fContent string) []byte {
	content := []byte{}
	content = append(content, byte(s))
	content = append(content, []byte(name)...)
	content = append(content, 00)
	content = append(content, []byte(fContent)...)
	return content
}
func makeHash(data []byte) string {
	sha := sha1.New()
	sha.Write(data)
	HEX := hex.EncodeToString(sha.Sum([]byte{}))
	return HEX
}
func makeObj(magicNumber int, content []byte) {
	sha := sha1.New()
	sha.Write([]byte{byte(magicNumber), 00})
	sha.Write([]byte((parseObj(content, magicNumber)).name))
	HEX := hex.EncodeToString(sha.Sum([]byte{}))
	err := os.Mkdir(".zil/objects/"+HEX[:2], 0755)
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write(content)
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	os.WriteFile(".zil/objects/"+HEX[:2]+"/"+HEX[2:], b.Bytes(), 0060)
	Err(err)
}

type obj struct {
	magicNumber int
	name        string
	content     string
}

func parseObj(raw []byte, magicNumber int) obj {
	headerNum := 0
	fName := ""
	content := ""
	for i, b := range raw {
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
	return obj{magicNumber: magicNumber, name: string(bytes.Trim([]byte(fName), "\x00")), content: content}
}
func readObj(sha string, magicNumber int) obj {
	path := ".zil/objects/" + sha[:2] + "/" + sha[2:]
	os.Chmod(path, 0755)
	defer os.Chmod(path, 0060)
	body, fErr := os.ReadFile(path)
	Err(fErr)
	gz, _ := gzip.NewReader(bytes.NewReader(body))
	d, bErr := ioutil.ReadAll(gz)
	Err(bErr)
	return parseObj(d, magicNumber)
}
func padNumberWithZero(value uint32) string {
	return fmt.Sprintf("%02d", value)
}
func appendTo(name, content string, MN int) {
	magicNumber := MN
	sha := sha1.New()
	sha.Write([]byte{byte(MN), 00})
	sha.Write([]byte(name))
	oldHEX := hex.EncodeToString(sha.Sum([]byte{}))
	magicNumber += 1
	sha.Reset()
	sha.Write([]byte{byte(magicNumber), 00})
	sha.Write([]byte(name))
	newHEX := hex.EncodeToString(sha.Sum([]byte{}))
	path := ".zil/objects/" + oldHEX[:2] + "/" + padNumberWithZero(uint32(magicNumber)) + newHEX[2:]
	diff := diffmatchpatch.New()
	magicNumber -= 1
	if magicNumber == 0 {
		//before := ".zil/objects" + oldHEX[:2] + "/"+ oldHEX[2:]
		obj := readObj(oldHEX, 0)
		d := diff.DiffMain(obj.content, content, true)
		fmt.Println(path)
		fmt.Println(d)
		os.WriteFile(path, assembleObjHeader(len(name), name, content), 0060)
	}
}

type pointerPair struct {
	path        string
	magicNumber int
	SHA         string
}

func getRoof(hash string) pointerPair {
	os.Chmod(".zil/ROOF", 0755)
	defer os.Chmod(".zil/ROOF", 0060)
	fCon, fErr := os.ReadFile(".zil/ROOF")
	Err(fErr)
	lines := strings.Split(string(fCon), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return pointerPair{}
	}
	results := []pointerPair{}
	for _, l := range lines {
		group := strings.Split(l, ":")
		if len(group) < 2 {
			continue
		}
		h := group[0]
		mn := group[1]
		if h == hash {
			r := pointerPair{}
			r.SHA = h
			intMN, intErr := strconv.Atoi(mn)
			Err(intErr)
			r.magicNumber = intMN
			r.path = ".zil/objects/" + h[:2] + "/"
			results = append(results, r)
		}
	}
	return results[len(results)-1]
}

func setRoof(name string, magicNumber int) {
	os.Chmod(".zil/ROOF", 0755)
	defer os.Chmod(".zil/ROOF", 0060)
	data := []byte{byte(magicNumber), 00}
	data = append(data, []byte(name)...)
	h := makeHash(data)
	bytes := []byte(h + ":" + strconv.Itoa(magicNumber) + "\n")
	f, err := os.OpenFile(".zil/ROOF", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	f.Write(bytes)
	Err(err)
}

func writeStage(dir string, magicNumber int, isGen bool) {
	fs.WalkDir(os.DirFS(dir), dir, func(path string, info fs.DirEntry, err error) error {
		Err(err)
		if info.IsDir() {
			return nil
		}
		if strings.Contains(path, ".zil") == true {
			return nil
		}
		fileContent, fErr := os.ReadFile(path)
		Err(fErr)
		data := []byte{byte(magicNumber), 00}
		data = append(data, []byte(info.Name())...)
		ROOFVAL := getRoof(makeHash(data))
		if ROOFVAL.path == "" {
			makeObj(0, assembleObjHeader(len(info.Name()), info.Name(), string(fileContent)))
			setRoof(info.Name(), 0)
		} else {
			appendTo(info.Name(), string(fileContent), ROOFVAL.magicNumber)
			setRoof(info.Name(), ROOFVAL.magicNumber+1)
		}
		return nil
	})
}
func main() {
	initZil()
	//content := []byte{}
	//content = append(content, []byte("hi.js")...)
	//content = append(content, 00)
	//content = append(content, []byte("HELLO woRLD!!")...)
	//makeObj(4, content)
	//ob := readObj("beaa9211a631fe91f8765c1ca4b29048f885ef9")
	//fmt.Println(ob)
	//writeStage(".", 0, true)
	writeStage(".", 0, false)
	//sha := sha1.New()
	//sha.Write([]byte{byte(0), 00})
	//sha.Write([]byte("main.go"))
	//HEX := hex.EncodeToString(sha.Sum([]byte{}))
	//fmt.Println(HEX)
}
