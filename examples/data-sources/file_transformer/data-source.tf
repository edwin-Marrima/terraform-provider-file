data "file_transformer" "name" {
  file                 = "./abc.json"
  override_array_items = false
  items = jsonencode(
    {
      "abc" = "newwww"
      "aaa" = ["a", "b"]
      c = {
        "name" = "Marrima"
      }
    }
  )
}