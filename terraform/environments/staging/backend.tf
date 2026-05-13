terraform {
  backend "s3" {
    bucket         = "hslee-terraform-state-ACCOUNT_ID"
    key            = "staging/terraform.tfstate"
    region         = "ap-northeast-2"
    dynamodb_table = "hslee-terraform-lock"
    encrypt        = true
  }
}
