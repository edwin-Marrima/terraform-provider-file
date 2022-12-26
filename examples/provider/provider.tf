terraform {
  required_providers {
    file = {
      version = "0.3.1"
      source  = "hashicorp.com/edu/file"
    }
  }
}
provider "file" {

}

data "file_transformer" "name" {
  file                 = "./.env"
  override_array_items = false
  items = <<EOT
	DB_PASSWORD=newpassword
	  USERNAME=marrima
    DB_NAME=sql
	EOT
}

data "file_transformer" "namex" {
  file                 = "./abcd.yml"
  override_array_items = true
 
  items = jsonencode(
          {
            "my-container" = {
                environment = ["NODE_ENV=production"]
            }
          }
  ) 
}