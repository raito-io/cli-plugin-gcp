module "demo" {
  count = var.demo_dataset ? 1 : 0

  source = "./demo"

  providers = {
    google = google
  }

  email_b_stewart   = var.email_b_stewart
  email_c_harris    = var.email_c_harris
  email_d_hayden    = var.email_d_hayden
  email_m_carissa   = var.email_m_carissa
  email_n_nguyen    = var.email_n_nguyen
  email_group_sales = var.email_group_sales
  email_group_dev   = var.email_group_dev
}

module "testing" {
  count = var.testing_dataset ? 1 : 0

  source = "./testing"

  providers = {
    google = google
  }

  email_b_stewart   = var.email_b_stewart
  email_c_harris    = var.email_c_harris
  email_d_hayden    = var.email_d_hayden
  email_m_carissa   = var.email_m_carissa
  email_n_nguyen    = var.email_n_nguyen
  email_group_sales = var.email_group_sales
  email_group_dev   = var.email_group_dev
}