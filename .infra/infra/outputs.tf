output "personas" {
  value = [
    var.email_b_stewart,
    var.email_c_harris,
    var.email_d_hayden,
    var.email_m_carissa,
    var.email_n_nguyen
  ]
}

output "project" {
  value = var.gcp_project
}

output "datasets" {
  value = concat(var.demo_dataset ? [module.demo[0].dataset] : [], var.testing_dataset ? [module.testing[0].dataset] : [])
}

output "tables" {
  value = [for each in [
    for x in [
      { active : var.demo_dataset, module : module.demo }, { active : var.testing_dataset, module : module.testing }
    ] : x.active ? { dataset : x.module[0].dataset, tables : x.module[0].tables } : null
  ] : each if each != null]
}