data "vantage_budgets" "all" {} 

output "all_budgets" {
  value = data.vantage_budgets.all
}