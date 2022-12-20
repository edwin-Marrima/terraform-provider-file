package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// TODO: explore test cleanup
// <Test-setup>
var testFile []TestFiles

type TestFiles struct {
	path    string
	content string
}

func init() {
	testFile = []TestFiles{
		{
			path:    "./test_artifact/init-file-001.json",
			content: `{"coach":"Mourinho"}`,
		},
	}
}
func TestMain(m *testing.M) {
	startUp()
	retCode := m.Run()
	//teardown
	tearDown()
	os.Exit(retCode)
}
func startUp() {
	for _, f := range testFile {
		file, _ := os.Create(f.path)
		defer file.Close()
		file.Write([]byte(f.content))
	}
}
func tearDown() {
	for _, f := range testFile {
		os.Remove(f.path)
	}
}

// </Test-setup>
func TestFolder(t *testing.T) {

	t.Run("Create folder when the one provided in the path does not exist", func(t *testing.T) {
		cl := Client{}
		filePath := []string{"./test_artifact/subfolder/file-0012.json"}
		for _, path := range filePath {
			_, err := cl.ReadHandler(path)
			fmt.Println("ERROR_CONTENT:", err)
			dirPath, _ := filepath.Split(path)
			assert.DirExists(t, dirPath)
			//clean test
			t.Cleanup(func() {
				os.RemoveAll(dirPath)
			})
		}
	})
	t.Run("Don't create folder when it already exists", func(t *testing.T) {

		cl := Client{}
		filePath := []string{"./test_artifact/subfolder/file-0013.json"}
		for _, path := range filePath {
			dirPath, _ := filepath.Split(path)
			//create folder
			os.MkdirAll(dirPath, 0777)
			_, err := cl.ReadHandler(path)
			fmt.Println("ERROR_CONTENT_@:", err)
			assert.DirExists(t, dirPath)

			//clean test
			t.Cleanup(func() {
				os.RemoveAll(dirPath)
			})
		}
	})
}

func TestFile(t *testing.T) {
	t.Run("Create file when the one provided in the path does not exist", func(t *testing.T) {

		cl := Client{}
		filePath := []string{"./test_artifact/file-001.json"}
		for _, path := range filePath {
			_, _ = cl.ReadHandler(path)
			assert.FileExists(t, path)
			//clean test
			t.Cleanup(func() {
				os.Remove(path)
			})
		}
	})
	t.Run("Don't overlap file content when provided file already exists", func(t *testing.T) {
		cl := Client{}
		filePath := []string{"./test_artifact/init-file-001.json"}
		for _, path := range filePath {
			_, _ = filepath.Split(path)
			fileT, _ := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
			defer fileT.Close()

			fileT.WriteAt([]byte(`{"coach":"Mourinho"}`), 0)

			fileData, err := os.ReadFile(path)
			b := make([]byte, 10)
			// read file
			fmt.Println("CCCTV:", err)

			file, _ := cl.ReadHandler(path)
			n, _ := file.Read(b)
			assert.Contains(t, string(fileData), string(b[:n]))
			if len(string(b[:n])) < 1 {
				t.Errorf("File is empty")
			}

		}
	})
}

//</UPDATE-FILE>

func TestFileTransformation(t *testing.T) {
	t.Run("Merge provided object with file content", func(t *testing.T) {
		testContent := []struct {
			cl                 Client
			srcContent         string
			filePath           string
			overrideArrayItems bool
			fileContent        map[string]interface{}
			expectedOutcome    map[string]interface{}
		}{
			{
				cl:                 Client{},
				srcContent:         `{"club":"Everton"}`,
				overrideArrayItems: false,
				filePath:           "./test_artifact/empty-source-001.json",
				fileContent: map[string]interface{}{
					"coach": "Mou",
					"Teams": []string{"Roma", "United", "Spurs", "Real", "Porto", "Inter"},
				},
				expectedOutcome: map[string]interface{}{
					"coach": "Mou",
					"Teams": []interface{}{"Roma", "United", "Spurs", "Real", "Porto", "Inter"},
					"club":  "Everton",
				},
			},
			{
				cl:                 Client{},
				srcContent:         `{"club":{"name":"Real Madrid"}}`,
				overrideArrayItems: false,
				filePath:           "./test_artifact/empty-source-001.json",
				fileContent: map[string]interface{}{
					"coach": "Mou",
					"club": map[string]interface{}{
						"stadium": "Bernabeu",
					},
				},
				expectedOutcome: map[string]interface{}{
					"coach": "Mou",
					"club": map[string]interface{}{
						"stadium": "Bernabeu",
						"name":    "Real Madrid",
					},
				},
			},
			{
				cl:                 Client{},
				srcContent:         `{"club":"Everton","Teams":["Porto", "Inter"]}`,
				overrideArrayItems: false,
				filePath:           "./test_artifact/empty-source-001.json",
				fileContent: map[string]interface{}{
					"coach": "Mou",
					"Teams": []string{"Roma", "United", "Spurs", "Real"},
				},
				expectedOutcome: map[string]interface{}{
					"club":  "Everton",
					"coach": "Mou",
					"Teams": []interface{}{"Roma", "United", "Spurs", "Real", "Porto", "Inter"},
				},
			},
			{
				cl:                 Client{},
				overrideArrayItems: true,
				srcContent:         `{"club":"Everton","Teams":["Porto", "Benfica"]}`,
				filePath:           "./test_artifact/empty-source-001.json",
				fileContent: map[string]interface{}{
					"Teams": []string{"Roma", "United", "Spurs", "Real"},
				},
				expectedOutcome: map[string]interface{}{
					"club":  "Everton",
					"Teams": []interface{}{"Porto", "Benfica"},
				},
			},
		}
		for _, value := range testContent {
			//Create file & register Content
			file, _ := os.OpenFile(value.filePath, os.O_CREATE|os.O_RDWR, 0666)
			b, _ := json.Marshal(value.fileContent)
			file.WriteAt(b, 0)
			file.Close()
			//assert phase

			value.cl.FileTransform(
				value.filePath,
				value.srcContent,
				value.filePath, WithOverrideArrayItems(value.overrideArrayItems),
			)
			//Reading the file content after running the function in order to obtain new file content
			actualFileContentInBytes, _ := os.ReadFile(value.filePath)
			actualFileContent := map[string]interface{}{}
			json.Unmarshal(actualFileContentInBytes, &actualFileContent)
			assert.Equal(t, value.expectedOutcome, actualFileContent)
			// Delete created file
			os.Remove(value.filePath)
		}
	})

	t.Run("Return permission error when provider has no permission to perform Read operation", func(t *testing.T) {
		//create file with no privileges
		cl := Client{}
		path := "./test_artifact/empty-source-002.json"
		file, _ := os.OpenFile(path, os.O_CREATE, 0000)
		file.Close()
		//assert phase
		err := cl.FileTransform(path, `{"name":"James"}`, filepath.Ext(path))
		assert.ErrorContains(t, err, "permission denied")

		//delete file
		file.Chmod(0777)
		os.Remove(path)
	})
	t.Run("Retun error when content of src or destination are invalid json object", func(t *testing.T) {
		testContent := []struct {
			cl                   Client
			srcContent           string
			filePath             string
			fileContent          string
			expectedErrorMessage string
		}{
			{
				cl:         Client{},
				srcContent: `{"club":{"name":"Real Madrid"cxxxx}}`,
				filePath:   "./test_artifact/empty-source-001.json",
				fileContent: `{
					"coach": "Mou",
					"club": {"stadium": "Bernabeu"}
				}`,
				expectedErrorMessage: "invalid character 'c' after object key:value pair",
			},
			{
				cl:         Client{},
				srcContent: `{"club":{"name":"Real Madrid"}}`,
				filePath:   "./test_artifact/empty-source-001.json",
				fileContent: `{
					"coach": "Mou",ba
					"club": {"stadium": "Bernabeu"}
				}`,
				expectedErrorMessage: "Content of file ./test_artifact/empty-source-001.json is malformed: invalid character 'b' looking for beginning of object key string",
			},
		}
		for _, value := range testContent {
			//Create file & register Content
			file, _ := os.OpenFile(value.filePath, os.O_CREATE|os.O_RDWR, 0666)
			file.WriteAt([]byte(value.fileContent), 0)
			file.Close()
			//assert phase

			err := value.cl.FileTransform(value.filePath, value.srcContent, value.filePath)
			assert.ErrorContains(t, err, value.expectedErrorMessage)
			// Delete created file
			os.Remove(value.filePath)
		}
	})
	t.Run("Create & insert items in nonexistent file", func(t *testing.T) {
		testContent := []struct {
			cl         Client
			srcContent string
			filePath   string
		}{
			{
				cl:         Client{},
				srcContent: `{"club":{"name":"Real Madrid"}}`,
				filePath:   "./test_artifact/nonexistent-001.json",
			},
		}
		for _, value := range testContent {
			value.cl.FileTransform(value.filePath, value.srcContent, value.filePath)

			srcContentMap := map[string]interface{}{}
			json.Unmarshal([]byte(value.srcContent), &srcContentMap)
			//assert
			actualFileContentInBytes, _ := os.ReadFile(value.filePath)
			actualFileContent := map[string]interface{}{}
			json.Unmarshal(actualFileContentInBytes, &actualFileContent)
			assert.Equal(t, srcContentMap, actualFileContent)
			// delete file
			os.Remove(value.filePath)
		}
	})

}

func TestYamlFileTransform(t *testing.T) {
	t.Run("Creates and register items in empty yaml file", func(t *testing.T) {
		testContent := []struct {
			cl         Client
			srcContent string
			filePath   string
		}{
			{
				cl:         Client{},
				filePath:   "./test_artifact/001.yaml",
				srcContent: `{"student-name": "Sagar"}`,
			},
			{
				cl:       Client{},
				filePath: "./test_artifact/002.yaml",
				srcContent: `{
					"coach": "Mou",
					"club": {"stadium": "Bernabeu"},
					"players":["Militao","Alaba","Nacho"]
				}`,
			},
		}
		for _, value := range testContent {
			value.cl.FileTransform(value.filePath, value.srcContent, value.filePath)
			//retrieve file content in order to test
			srcContentMap := map[string]interface{}{}
			json.Unmarshal([]byte(value.srcContent), &srcContentMap)

			actualFileContentInBytes, _ := os.ReadFile(value.filePath)
			actualFileContent := map[string]interface{}{}
			yaml.Unmarshal(actualFileContentInBytes, &actualFileContent)
			assert.Equal(t, srcContentMap, actualFileContent)
			// delete file
			os.Remove(value.filePath)
		}
	})
	t.Run("Replace content in previously created (existent) yaml file", func(t *testing.T) {
		testContent := []struct {
			cl              Client
			srcContent      string
			filePath        string
			fileContent     map[string]interface{}
			expectedOutcome map[string]interface{}
		}{
			{
				cl:         Client{},
				srcContent: `{"club":"Roma"}`,
				filePath:   "./test_artifact/replace-001.yml",
				fileContent: map[string]interface{}{
					"club":  "Sevilha",
					"Teams": []string{"Roma", "United", "Spurs", "Real", "Porto", "Inter"},
				},
				expectedOutcome: map[string]interface{}{
					"Teams": []interface{}{"Roma", "United", "Spurs", "Real", "Porto", "Inter"},
					"club":  "Roma",
				},
			},
			{
				cl:         Client{},
				srcContent: `{"Teams":["Bayern"]}`,
				filePath:   "./test_artifact/replace-001.yml",
				fileContent: map[string]interface{}{
					"club":  "Sevilha",
					"Teams": []string{"Roma", "United", "Spurs", "Real", "Porto", "Inter"},
				},
				expectedOutcome: map[string]interface{}{
					"Teams": []interface{}{"Roma", "United", "Spurs", "Real", "Porto", "Inter", "Bayern"},
					"club":  "Sevilha",
				},
			},
		}
		for _, value := range testContent {
			//Create file & register Content
			file, _ := os.OpenFile(value.filePath, os.O_CREATE|os.O_RDWR, 0666)
			b, _ := json.Marshal(value.fileContent)
			file.WriteAt(b, 0)
			file.Close()
			//assert phase

			value.cl.FileTransform(value.filePath, value.srcContent, value.filePath)
			//Reading the file content after running the function in order to obtain new file content
			actualFileContentInBytes, _ := os.ReadFile(value.filePath)

			actualFileContent := map[string]interface{}{}
			yaml.Unmarshal(actualFileContentInBytes, &actualFileContent)
			assert.Equal(t, value.expectedOutcome, actualFileContent)
			// Delete created file
			os.Remove(value.filePath)
		}
	})
}

func TestConvertFileJsonYaml(t *testing.T) {
	t.Run("Convert files (yaml to json) & ", func(t *testing.T) {
		testContent := []struct {
			cl                  Client
			srcContent          string
			filePath            string
			fileContent         string
			outputPath          string
			expectedFileContent map[string]interface{}
		}{
			{
				cl:         Client{},
				srcContent: `{"club":{"name":"Real Madrid"}}`,
				filePath:   "./test_artifact/convert-001.json",
				fileContent: `{
					"coach": "Mou",
					"club": {"stadium": "Bernabeu"}
				}`,
				outputPath: "./test_artifact/convert-yaml-001.yml",
				expectedFileContent: map[string]interface{}{
					"coach": "Mou",
					"club": map[string]interface{}{"stadium": "Bernabeu",
						"name": "Real Madrid",
					},
				},
			},
			{
				cl:         Client{},
				srcContent: `{"club":{"name":"Real Madrid"}}`,
				filePath:   "./test_artifact/convert-003.yml",
				fileContent: `{
					"coach": "Mou",
					"club": {"stadium": "Bernabeu"}
				}`,
				outputPath: "./test_artifact/convert-json-003.yml",
				expectedFileContent: map[string]interface{}{
					"coach": "Mou",
					"club": map[string]interface{}{"stadium": "Bernabeu",
						"name": "Real Madrid",
					},
				},
			},
		}
		for _, value := range testContent {
			//Create file & register Content
			file, _ := os.OpenFile(value.filePath, os.O_CREATE|os.O_RDWR, 0777)
			file.WriteAt([]byte(value.fileContent), 0)
			file.Close()
			//assert phase

			value.cl.FileTransform(value.filePath, value.srcContent, value.outputPath)
			assert.FileExists(t, value.outputPath)
			// Delete created file
			os.Remove(value.filePath)
			os.Remove(value.outputPath)
		}
	})
}

//<ENV FILE>

func TestEnvFileEdit(t *testing.T) {
	t.Run("Add keys in .env file", func(t *testing.T) {
		testContent := []struct {
			cl                  Client
			filePath            string
			fileContent         string
			newEnvContent       string
			expectedFileContent map[string]string
		}{
			{
				cl:       Client{},
				filePath: "./test_artifact/.env",
				fileContent: `
				DB_HOST=localhost
				DB_USER=admin
				DB_PASSWORD=password
				`,
				newEnvContent: `
				VERSION=1.1.2
				`,
				expectedFileContent: map[string]string{
					"DB_HOST":     "localhost",
					"DB_USER":     "admin",
					"DB_PASSWORD": "password",
					"VERSION":     "1.1.2",
				},
			},
			{
				cl:       Client{},
				filePath: "./test_artifact/.env",
				fileContent: `
				DB_HOST=localhost
				DB_USER=admin
				DB_PASSWORD=password
				`,
				newEnvContent: `
				DB_PASSWORD=newpassword
				`,
				expectedFileContent: map[string]string{
					"DB_HOST":     "localhost",
					"DB_USER":     "admin",
					"DB_PASSWORD": "newpassword",
				},
			},
			// ignore commented files
			{
				cl:       Client{},
				filePath: "./test_artifact/.env",
				fileContent: `
				#DB_HOST=localhost
				DB_USER=admin
				DB_PASSWORD=password
				`,
				newEnvContent: `
				DB_PASSWORD=newpassword
				`,
				expectedFileContent: map[string]string{
					"DB_USER":     "admin",
					"DB_PASSWORD": "newpassword",
				},
			},
		}
		for _, value := range testContent {
			//Create file & register Content
			file, _ := os.OpenFile(value.filePath, os.O_CREATE|os.O_RDWR, 0777)
			file.WriteAt([]byte(value.fileContent), 0)
			file.Close()

			value.cl.FileTransform(value.filePath, value.newEnvContent, value.filePath)
			//retrieve new .env content
			b, _ := os.ReadFile(value.filePath)
			envFile, _ := godotenv.Unmarshal(string(b))
			assert.Equal(t, value.expectedFileContent, envFile)
			// Delete created file
			os.Remove(value.filePath)
		}
	})
}
