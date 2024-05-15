variable "gcp_project" {
  type      = string
  nullable  = false
  sensitive = false
}

variable "gcp_region" {
  type      = string
  nullable  = false
  sensitive = false
}

variable "email_b_stewart" {
  type      = string
  nullable  = false
  sensitive = false
  default   = "b_stewart@raito.dev"
}

variable "email_c_harris" {
  type      = string
  nullable  = false
  sensitive = false
  default   = "c_harris@raito.dev"
}

variable "email_d_hayden" {
  type      = string
  nullable  = false
  sensitive = false
  default   = "d_hayden@raito.dev"
}

variable "email_m_carissa" {
  type      = string
  nullable  = false
  sensitive = false
  default   = "m_carissa@raito.dev"
}

variable "email_n_nguyen" {
  type      = string
  nullable  = false
  sensitive = false
  default   = "n_nguyen@raito.dev"
}

variable "email_group_sales" {
  type      = string
  nullable  = false
  sensitive = false
  default   = "sales@raito.dev"
}

variable "email_group_dev" {
  type      = string
  nullable  = false
  sensitive = false
  default   = "dev@raito.dev"
}