terraform {
  required_providers {
    pwpusher = {
      source = "hashicorp.com/edu/pwpusher"
    }
  }
}
provider "pwpusher" {
  # example configuration here
  url = "http://localhost:5100"
}
