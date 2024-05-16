resource "google_project_iam_binding" "project_editor" {
  project = var.gcp_project
  role    = "roles/editor"

  members = [
    "user:${var.email_n_nguyen}",
  ]
}

resource "google_project_iam_binding" "project_bigquery_job_user" {
  project = var.gcp_project
  role    = "roles/bigquery.jobUser"

  members = [
    "user:${var.email_b_stewart}",
    "user:${var.email_c_harris}",
    "user:${var.email_d_hayden}",
    "user:${var.email_m_carissa}",
    "user:${var.email_n_nguyen}",
  ]
}