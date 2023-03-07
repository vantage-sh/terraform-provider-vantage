variable "aws_assume_role" {
  type = string
}

variable "vantage_sns_topic" {
  type    = string
  default = "arn:aws:sns:us-east-1:630399649041:cost-and-usage-report-uploaded"
}
