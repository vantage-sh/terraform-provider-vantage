resource "vantage_mongodb_provider" "example" {
  cluster_uri = "mongodb+srv://cluster0.mongodb.net/test"
  api_key    = "supersecretapikey"
}