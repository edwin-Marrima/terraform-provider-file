package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Client struct {
}

type Transformer struct {
	path               string
	outputPath         string
	items              string
	overrideArrayItems bool
}

func WithOverrideArrayItems(append bool) func(*Transformer) {
	return func(m *Transformer) {
		m.overrideArrayItems = append
	}
}

type Unmarshal func(in []byte, out interface{}) (err error)
type Marshal func(in interface{}) (out []byte, err error)

type DataDecoder struct {
	unmarshal Unmarshal
	marshal   Marshal
}

var (
	supportedFileExtDecode = map[string]Unmarshal{
		".yaml": yaml.Unmarshal,
		".yml":  yaml.Unmarshal,
		".json": json.Unmarshal,
	}
	supportedFileExtEncode = map[string]Marshal{
		".yaml": yaml.Marshal,
		".yml":  yaml.Marshal,
		".json": json.Marshal,
	}
)

func (cl Client) FileTransform(path, content, outputPath string, options ...func(*Transformer)) error {
	t := Transformer{path: path, items: content, outputPath: outputPath, overrideArrayItems: false}
	for _, opt := range options {
		opt(&t)
	}
	file, err := cl.ReadHandler(t.path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	// create slice that hold file content in bytes, with size equal to file size
	b := make([]byte, fileInfo.Size())
	_, err = file.Read(b)
	if err != nil {
		return err
	}
	if ok, _ := regexp.MatchString(".env", filepath.Ext(path)); ok {
		return cl.dotEnv(b, path, content)
	}
	return cl.jsonAndYaml(b, t)
}
func (cl Client) jsonAndYaml(b []byte, t Transformer) error {

	dstContent := map[string]interface{}{}
	srcContent := map[string]interface{}{}
	dataDecoder := DataDecoder{
		unmarshal: supportedFileExtDecode[filepath.Ext(t.path)],
		marshal:   supportedFileExtEncode[filepath.Ext(t.outputPath)],
	}

	// Unmarshal empty json/map (empty byte array=>b=0) we will get 'unexpected end of JSON input' error
	//The conditional below aims to workaround this error
	if len(b) > 0 {
		err := dataDecoder.unmarshal(b, &dstContent)
		if err != nil {
			return errors.New(fmt.Sprintf("Content of file %s is malformed: %s", t.path, err.Error()))
		}
	}
	err := json.Unmarshal([]byte(t.items), &srcContent)
	if err != nil {
		return err
	}
	mergedContent, err := Merge(srcContent, dstContent, WithOverrideArray(t.overrideArrayItems))
	if err != nil {
		return err
	}

	var mergedContentB []byte

	mergedContentB, err = dataDecoder.marshal(mergedContent)
	if err != nil {
		return err
	}

	fileWriteP, err := os.OpenFile(t.outputPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	fileWriteP.Truncate(0)
	if err != nil {
		return err
	}
	defer fileWriteP.Close()

	// replace all content of file with merged content
	_, err = fileWriteP.Write(mergedContentB)

	if err != nil {
		return err
	}
	fileWriteP.Sync()
	return nil
}

func (cl Client) dotEnv(b []byte, path, content string) error {

	fileContent, err := godotenv.Unmarshal(string(b))
	if err != nil {
		return err
	}
	envMap, err := godotenv.Unmarshal(content)
	if err != nil {
		return err
	}
	// merging environment variables to map that contains provided file (.env) environment variables
	for k, v := range envMap {
		fileContent[k] = v
	}
	err = godotenv.Write(fileContent, path)
	if err != nil {
		return err
	}
	return nil
}

func (cl Client) ReadHandler(path string) (*os.File, error) {
	dirPath, _ := filepath.Split(path)
	fmt.Println("AAAAAAAAAAAAA:" + dirPath)
	// check if directory exists and create new one if not
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0777)
	}

	//If the file does not exist, a new file is created.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)

	if err != nil {
		return nil, err
	}
	return file, nil
}
