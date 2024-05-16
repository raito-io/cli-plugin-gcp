locals {
  sales_tables = {
    "sales_country_region_currency"         = google_bigquery_table.Sales_CountryRegionCurrency,
    "sales_credit_card"                     = google_bigquery_table.Sales_CreditCard,
    "sales_currency"                        = google_bigquery_table.Sales_Currency,
    "sales_currency_rate"                   = google_bigquery_table.Sales_CurrencyRate,
    "sales_customer"                        = google_bigquery_table.Sales_Customer,
    "sales_person_credit_card"              = google_bigquery_table.Sales_PersonCreditCard,
    "sales_sales_order_detail"              = google_bigquery_table.Sales_SalesOrderDetail,
    "sales_sales_order_header"              = google_bigquery_table.Sales_SalesOrderHeader,
    "sales_sales_order_header_sales_reason" = google_bigquery_table.Sales_SalesOrderHeaderSalesReason,
    "sales_sales_person"                    = google_bigquery_table.Sales_SalesPerson,
    "sales_sales_person_quota_history"      = google_bigquery_table.Sales_SalesPersonQuotaHistory,
    "sales_sales_reason"                    = google_bigquery_table.Sales_SalesReason,
    "sales_sales_tax_rate"                  = google_bigquery_table.Sales_SalesTaxRate,
    "sales_sales_territory"                 = google_bigquery_table.Sales_SalesTerritory,
    "sales_sales_territory_history"         = google_bigquery_table.Sales_SalesTerritoryHistory,
    "sales_shopping_cart_item"              = google_bigquery_table.Sales_ShoppingCartItem,
    "sales_special_offer"                   = google_bigquery_table.Sales_SpecialOffer,
    "sales_special_offer_product"           = google_bigquery_table.Sales_SpecialOfferProduct,
    "sales_store"                           = google_bigquery_table.Sales_Store
  }

  human_resource_tables = {
    "human_resources_department"                  = google_bigquery_table.HumanResources_Department,
    "human_resources_employee"                    = google_bigquery_table.HumanResources_Employee,
    "human_resources_employee_department_history" = google_bigquery_table.HumanResources_EmployeeDepartmentHistory,
    "human_resources_employee_pay_history"        = google_bigquery_table.HumanResources_EmployeePayHistory,
    "human_resources_job_candidate"               = google_bigquery_table.HumanResources_JobCandidate,
    "human_resources_shift"                       = google_bigquery_table.HumanResources_Shift,
  }
}

resource "google_bigquery_table_iam_binding" "bq_data_viewer_sales" {
  for_each = local.sales_tables

  dataset_id = each.value.dataset_id
  table_id   = each.value.table_id
  role       = "roles/bigquery.dataViewer"
  members = [
    "user:${var.email_m_carissa}",
    "user:${var.email_d_hayden}",
    "group:${var.email_group_sales}",
    "group:${var.email_group_dev}"
  ]
}

resource "google_bigquery_table_iam_binding" "bq_data_viewer_human_resources" {
  for_each   = local.human_resource_tables
  dataset_id = each.value.dataset_id
  table_id   = each.value.table_id
  role       = "roles/bigquery.dataViewer"
  members = [
    "user:${var.email_m_carissa}",
  ]
}

resource "google_bigquery_dataset_iam_binding" "db_dataset_binding" {
  dataset_id = google_bigquery_dataset.dataset.dataset_id
  members = [
    "user:${var.email_d_hayden}"
  ]
  role = "roles/bigquery.dataViewer"
}