resource "vantage_aws_provider" "example" {
  cross_account_arn = "arn:aws:iam::123456789012:role/CrossAccountRole"
  bucket_arn        = "arn:aws:s3:::my-cur-bucket"
}