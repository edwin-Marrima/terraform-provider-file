package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/joho/godotenv"
)

//<properties-validation>

func TestAccReturnErrorWhenInputAndOutputFileExtensionInNotSupported(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//invalid output file extension
			{
				Config: `
				data "file_transformer" "foo" {
					file = "./test_assets/file-000.json"
					output = "./test_assets/file-000.xml"
					items = ""
				}
				`,
				ExpectError: regexp.MustCompile("The file extension is not supported"),
			},
			//invalid input file extension
			{
				Config: `
				data "file_transformer" "foo" {
					file = "./test_assets/file-000.xml"
					output = "./test_assets/file-000.json"
					items = ""
				}
				`,
				ExpectError: regexp.MustCompile("The file extension is not supported"),
			},
		},
	})
}

//</properties-validation>
func TestAccRegisterElementsWithEmptyFile(t *testing.T) {
	filePath := "./test_assets/file-001.json"

	resource.Test(t, resource.TestCase{

		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(*terraform.State) error {
			os.Remove(filePath)
			return nil
		},
		Steps: []resource.TestStep{
			{

				Config: testAccCheckFileTransformConfig(filePath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.file_transformer.foo", "file", filePath),
					resource.TestCheckResourceAttr("data.file_transformer.foo", "output", filePath),
					testAccCheckFileTransformerExists("data.file_transformer.foo"),
					testAccCheckFileTransformerCreatedFile("data.file_transformer.foo"),
				),
			},
			{
				Config: testAccCheckFileTransformConfigArrayOverride(filePath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileTransformerCreatedFile("data.file_transformer.foo"),
				),
			},
		},
	})
}

func testAccCheckFileTransformConfigArrayOverride(filePath string) string {
	return fmt.Sprintf(`
		data "file_transformer" "foo" {
			file = "%s"
			override_array_items = true
			items = jsonencode(
				{
					"abc"="new"
					"aaa" = [12,14]
						c = {
							"name"="edwin"
						}
					}
				)
		}
	`, filePath)
}
func testAccCheckFileTransformConfig(filePath string) string {
	return fmt.Sprintf(`
		data "file_transformer" "foo" {
			file = "%s"
			override_array_items = true
			items = jsonencode(
				{
					"abc"="new"
					"aaa" = ["aa","bb","12"]
						c = {
							"name"="edwin"
						}
					}
				)
		}
	`, filePath)
}

//<file-conversion>
func TestAccFileConversion(t *testing.T) {
	filePath := "./test_assets/file-002.json"
	outputFilePath := "./test_assets/file-003.yaml"
	initFileContent := map[string]interface{}{"name": "Lyon", "players": []string{"Tolisso", "Lacazette", "Fekir"}}
	b, _ := json.Marshal(initFileContent)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			os.Remove(filePath)
			os.Remove(outputFilePath)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					//Create file & register Content
					file, _ := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
					file.WriteAt(b, 0)
					file.Close()
				},
				Config: testAccCheckFileConvertFileFormat(filePath, outputFilePath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileTransformerCreatedFile("data.file_transformer.fooo"),
					// testAccFileTransformerContent("data.file_transformer.fooo", initFileContent),
				),
			},
		},
	})
}

func testAccCheckFileConvertFileFormat(filePath, outputFilePath string) string {
	return fmt.Sprintf(`
		data "file_transformer" "fooo" {
			file = "%s"
			output = "%s"
			override_array_items = true
			items = jsonencode({})
		}
	`, filePath, outputFilePath)
}

//</file-conversion>

//<.env file>
func TestAccDotEnvFileEdit(t *testing.T) {
	filePath := "./test_assets/.env"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNewDotEnvFile(filePath, "DB_PASSWORD=newpassword\nUSERNAME=user123"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileTransformerCreatedFile("data.file_transformer.fooo2"),
					testAccCheckDotEnvFileContent("data.file_transformer.fooo2", []string{"DB_PASSWORD", "USERNAME"}),
				),
			},
		},
	})
}
func testAccNewDotEnvFile(filePath, items string) string {
	return fmt.Sprintf(`
		data "file_transformer" "fooo2" {
			file = "%s"
			items = <<EOT
			DB_PASSWORD=newpassword
			USERNAME=user123
			EOT
		}
	`, filePath)
}
func testAccCheckDotEnvFileContent(n string, keys []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//ready state file
		rs, _ := s.RootModule().Resources[n]
		filePath := rs.Primary.Attributes["file"]
		// read file content
		b, _ := os.ReadFile(filePath)
		//transform .env file content in map[string]string
		fileContent, _ := godotenv.Unmarshal(string(b))
		for _, k := range keys {
			if _, ok := fileContent[k]; !ok {
				return errors.New(fmt.Sprintf("Key `%s` does not exist in %s", k, filePath))
			}
		}
		return nil
	}
}

//</.env file>

func testAccCheckFileTransformerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set")
		}
		return nil
	}
}
func testAccCheckFileTransformerCreatedFile(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, _ := s.RootModule().Resources[n]
		filePath := rs.Primary.Attributes["file"]
		//reading, in order to verify if its was created or not
		_, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccFileTransformerContent(n string, expectedFileItems map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, _ := s.RootModule().Resources[n]
		filePath := rs.Primary.Attributes["file"]
		//reading, in order to verify if its was created or not
		b, _ := os.ReadFile(filePath)
		var fileItems map[string]interface{}
		json.Unmarshal(b, &fileItems)
		if !reflect.DeepEqual(fileItems, expectedFileItems) {
			return errors.New(fmt.Sprintf("Content of file %s is' equal to %v whereas the content provided is %v", filePath, fileItems, expectedFileItems))
		}
		return nil
	}
}
