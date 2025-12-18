resource "vantage_gcp_provider" "example" {
  project_id      = "my-gcp-project"
  billing_account = "000000-111111-222222"
  service_account = <<EOF
{
  "type": "service_account",
  "project_id": "my-gcp-project",
  "private_key_id": "xxxxxxx",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "terraform@my-gcp-project.iam.gserviceaccount.com",
  "client_id": "9876543210",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/terraform@my-gcp-project.iam.gserviceaccount.com"
}
EOF
}