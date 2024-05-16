output "dataset" {
  value = google_bigquery_dataset.dataset.dataset_id
}

output "tables" {
  value = [
    google_bigquery_table.Sales_CountryRegionCurrency.table_id,
    google_bigquery_table.Sales_CreditCard.table_id,
    google_bigquery_table.Sales_Currency.table_id,
    google_bigquery_table.Sales_CurrencyRate.table_id,
    google_bigquery_table.Sales_Customer.table_id,
    google_bigquery_table.Sales_PersonCreditCard.table_id,
    google_bigquery_table.Sales_SalesOrderDetail.table_id,
    google_bigquery_table.Sales_SalesOrderHeader.table_id,
    google_bigquery_table.Sales_SalesOrderHeaderSalesReason.table_id,
    google_bigquery_table.Sales_SalesPerson.table_id,
    google_bigquery_table.Sales_SalesPersonQuotaHistory.table_id,
    google_bigquery_table.Sales_SalesReason.table_id,
    google_bigquery_table.Sales_SalesTaxRate.table_id,
    google_bigquery_table.Sales_SalesTerritory.table_id,
    google_bigquery_table.Sales_SalesTerritoryHistory.table_id,
    google_bigquery_table.Sales_ShoppingCartItem.table_id,
    google_bigquery_table.Sales_SpecialOffer.table_id,
    google_bigquery_table.Sales_SpecialOfferProduct.table_id,
    google_bigquery_table.Sales_Store.table_id,
    google_bigquery_table.HumanResources_Department.table_id,
    google_bigquery_table.HumanResources_Employee.table_id,
    google_bigquery_table.HumanResources_EmployeeDepartmentHistory.table_id,
    google_bigquery_table.HumanResources_EmployeePayHistory.table_id,
    google_bigquery_table.HumanResources_JobCandidate.table_id,
    google_bigquery_table.HumanResources_Shift.table_id
  ]
}