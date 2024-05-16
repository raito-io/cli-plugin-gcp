resource "google_bigquery_dataset" "dataset" {
  dataset_id                 = "MASTER_DATA"
  location                   = "EU"
  delete_contents_on_destroy = true
}

resource "google_data_catalog_taxonomy" "business_criticality" {
  display_name           = "Business Criticality"
  activated_policy_types = ["FINE_GRAINED_ACCESS_CONTROL"]
  region                 = "eu"
}

resource "google_data_catalog_policy_tag" "business_criticality_high" {
  taxonomy     = google_data_catalog_taxonomy.business_criticality.id
  display_name = "High"
}

resource "google_data_catalog_policy_tag" "business_criticality_high_employee_snn" {
  taxonomy          = google_data_catalog_taxonomy.business_criticality.id
  display_name      = "employee_snn"
  parent_policy_tag = google_data_catalog_policy_tag.business_criticality_high.id
}

resource "google_data_catalog_policy_tag" "business_criticality_medium" {
  taxonomy     = google_data_catalog_taxonomy.business_criticality.id
  display_name = "Medium"
}

resource "google_data_catalog_policy_tag" "business_ciritcality_medium_age" {
  taxonomy          = google_data_catalog_taxonomy.business_criticality.id
  display_name      = "age"
  parent_policy_tag = google_data_catalog_policy_tag.business_criticality_medium.id
}

resource "google_bigquery_table" "dbo_DatabaseLog" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "dbo_DatabaseLog"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "name": "SystemInformationID",
    "type": "INTEGER",
    "mode": "NULLABLE"
  },
  {
    "name": "Database_Version",
    "type": "STRING",
    "mode": "NULLABLE"
  },
  {
    "name": "VersionDate",
    "type": "TIMESTAMP",
    "mode": "NULLABLE"
  },
  {
    "name": "ModifiedDate",
    "type": "TIMESTAMP",
    "mode": "NULLABLE"
  }
]
EOF

}

resource "google_bigquery_table" "dbo_AWBuildVersion" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "dbo_AWBuildVersion"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "DatabaseLogID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PostTime",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "DatabaseUser",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Event",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Schema",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Object",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "TSQL",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "XmlEvent",
    "type": "STRING"
  }
]
EOF

}

resource "google_bigquery_table" "dbo_ErrorLog" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "dbo_ErrorLog"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "ErrorLogID",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ErrorTime",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "UserName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ErrorNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ErrorSeverity",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ErrorState",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ErrorProcedure",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ErrorLine",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ErrorMessage",
    "type": "STRING"
  }
]
EOF
}

resource "google_bigquery_table" "HumanResources_Department" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "HumanResources_Department"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "DepartmentID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "GroupName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "HumanResources_Employee" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "HumanResources_Employee"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "NationalIDNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "LoginID",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "OrganizationNode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "OrganizationLevel",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "JobTitle",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "BirthDate",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "MaritalStatus",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Gender",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "HireDate",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "SalariedFlag",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "VacationHours",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "SickLeaveHours",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "CurrentFlag",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "HumanResources_EmployeeDepartmentHistory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "HumanResources_EmployeeDepartmentHistory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "DepartmentID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ShiftID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StartDate",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "EndDate",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "HumanResources_EmployeePayHistory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "HumanResources_EmployeePayHistory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "RateChangeDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "Rate",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "PayFrequency",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "HumanResources_JobCandidate" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "HumanResources_JobCandidate"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "JobCandidateID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Resume",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "HumanResources_Shift" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "HumanResources_Shift"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "ShiftID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "StartTime",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "EndTime",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_Address" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_Address"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "AddressID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "AddressLine1",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "AddressLine2",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "City",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "StateProvinceID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PostalCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "SpatialLocation",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_AddressType" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_AddressType"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "AddressTypeID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_BusinessEntity" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_BusinessEntity"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_BusinessEntityAddress" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_BusinessEntityAddress"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "AddressID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "AddressTypeID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_BusinessEntityContact" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_BusinessEntityContact"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PersonID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ContactTypeID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_ContactType" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_ContactType"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "ContactTypeID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_CountryRegion" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_CountryRegion"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "CountryRegionCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_EmailAddress" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_EmailAddress"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "policyTags": {
      "names": [
        "${google_data_catalog_policy_tag.business_criticality_high_employee_snn.name}"
      ]
    },
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "EmailAddressID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "EmailAddress",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_Password" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_Password"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PasswordHash",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "PasswordSalt",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_Person" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_Person"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PersonType",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "NameStyle",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Title",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "FirstName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "MiddleName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "LastName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Suffix",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "EmailPromotion",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "AdditionalContactInfo",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Demographics",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_PersonPhone" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_PersonPhone"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PhoneNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "PhoneNumberTypeID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_PhoneNumberType" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_PhoneNumberType"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "PhoneNumberTypeID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Person_StateProvince" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Person_StateProvince"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "StateProvinceID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StateProvinceCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "CountryRegionCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "IsOnlyStateProvinceFlag",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "TerritoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_BillOfMaterials" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_BillOfMaterials"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BillOfMaterialsID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductAssemblyID",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ComponentID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "EndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "UnitMeasureCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "BOMLevel",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PerAssemblyQty",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_Culture" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_Culture"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "CultureID",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_Document" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_Document"
  deletion_protection = false
  schema              = <<EOF
 [
  {
    "mode": "NULLABLE",
    "name": "DocumentNode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "DocumentLevel",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Title",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Owner",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "FolderFlag",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "FileName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "FileExtension",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Revision",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ChangeNumber",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Status",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "DocumentSummary",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Document",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_Illustration" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_Illustration"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "IllustrationID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Diagram",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_Location" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_Location"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "LocationID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "CostRate",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Availability",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_Product" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_Product"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "MakeFlag",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "FinishedGoodsFlag",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Color",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "SafetyStockLevel",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ReorderPoint",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StandardCost",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ListPrice",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Size",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "SizeUnitMeasureCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "WeightUnitMeasureCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Weight",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "DaysToManufacture",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductLine",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Class",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Style",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductSubcategoryID",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductModelID",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "SellStartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "SellEndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "DiscontinuedDate",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductCategory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductCategory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductCategoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductCostHistory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductCostHistory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "EndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "StandardCost",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductDescription" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductDescription"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductDescriptionID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Description",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductDocument" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductDocument"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "DocumentNode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductInventory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductInventory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "LocationID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Shelf",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Bin",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Quantity",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductListPriceHistory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductListPriceHistory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "EndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ListPrice",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductModel" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductModel"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductModelID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "CatalogDescription",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Instructions",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductModelIllustration" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductModelIllustration"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductModelID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "IllustrationID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductModelProductDescriptionCulture" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductModelProductDescriptionCulture"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductModelID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductDescriptionID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "CultureID",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductPhoto" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductPhoto"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductPhotoID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ThumbNailPhoto",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ThumbnailPhotoFileName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "LargePhoto",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "LargePhotoFileName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductProductPhoto" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductProductPhoto"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductPhotoID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Primary",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductReview" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductReview"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductReviewID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ReviewerName",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ReviewDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "EmailAddress",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Rating",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Comments",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ProductSubcategory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ProductSubcategory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductSubcategoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductCategoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_ScrapReason" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_ScrapReason"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ScrapReasonID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_TransactionHistory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_TransactionHistory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "TransactionID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ReferenceOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ReferenceOrderLineID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "TransactionDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "TransactionType",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Quantity",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ActualCost",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_TransactionHistoryArchive" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_TransactionHistoryArchive"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "TransactionID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ReferenceOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ReferenceOrderLineID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "TransactionDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "TransactionType",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Quantity",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ActualCost",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_UnitMeasure" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_UnitMeasure"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "UnitMeasureCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Production_WorkOrder" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_WorkOrder"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "WorkOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "OrderQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StockedQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ScrappedQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "EndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "DueDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ScrapReasonID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
  EOF
}

resource "google_bigquery_table" "Production_WorkOrderRouting" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Production_WorkOrderRouting"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "WorkOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "OperationSequence",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "LocationID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ScheduledStartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ScheduledEndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ActualStartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ActualEndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ActualResourceHrs",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "PlannedCost",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ActualCost",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Purchasing_ProductVendor" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Purchasing_ProductVendor"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "AverageLeadTime",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StandardPrice",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "LastReceiptCost",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "LastReceiptDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "MinOrderQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "MaxOrderQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "OnOrderQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "UnitMeasureCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Purchasing_PurchaseOrderDetail" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Purchasing_PurchaseOrderDetail"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "PurchaseOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "PurchaseOrderDetailID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "DueDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "OrderQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "UnitPrice",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "LineTotal",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ReceivedQty",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "RejectedQty",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "StockedQty",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Purchasing_PurchaseOrderHeader" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Purchasing_PurchaseOrderHeader"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "PurchaseOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "RevisionNumber",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Status",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "EmployeeID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "VendorID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ShipMethodID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "OrderDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ShipDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "SubTotal",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "TaxAmt",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Freight",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "TotalDue",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Purchasing_ShipMethod" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Purchasing_ShipMethod"
  deletion_protection = false
  schema              = <<EOF
[
  {
    "mode" : "NULLABLE",
    "name" : "ShipMethodID",
    "type" : "INTEGER"
  },
  {
    "mode" : "NULLABLE",
    "name" : "Name",
    "type" : "STRING"
  },
  {
    "mode" : "NULLABLE",
    "name" : "ShipBase",
    "type" : "FLOAT"
  },
  {
    "mode" : "NULLABLE",
    "name" : "ShipRate",
    "type" : "FLOAT"
  },
  {
    "mode" : "NULLABLE",
    "name" : "rowguid",
    "type" : "STRING"
  },
  {
    "mode" : "NULLABLE",
    "name" : "ModifiedDate",
    "type" : "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Purchasing_Vendor" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Purchasing_Vendor"
  deletion_protection = false
  schema              = <<EOF
    [
    {
        "mode" : "NULLABLE",
        "name" : "BusinessEntityID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "AccountNumber",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "Name",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "CreditRating",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "PreferredVendorStatus",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ActiveFlag",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "PurchasingWebServiceURL",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ModifiedDate",
        "type" : "TIMESTAMP"
    }
    ]
EOF
}

resource "google_bigquery_table" "Sales_CountryRegionCurrency" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_CountryRegionCurrency"
  deletion_protection = false
  schema              = <<EOF
    [
    {
        "mode" : "NULLABLE",
        "name" : "CountryRegionCode",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "CurrencyCode",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ModifiedDate",
        "type" : "TIMESTAMP"
    }
    ]
EOF
}

resource "google_bigquery_table" "Sales_CreditCard" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_CreditCard"
  deletion_protection = false
  schema              = <<EOF
    [
    {
        "mode" : "NULLABLE",
        "name" : "CreditCardID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "CardType",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "CardNumber",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ExpMonth",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ExpYear",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ModifiedDate",
        "type" : "TIMESTAMP"
    }
    ]
  EOF
}

resource "google_bigquery_table" "Sales_Currency" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_Currency"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "CurrencyCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_CurrencyRate" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_CurrencyRate"
  deletion_protection = false
  schema              = <<EOF
    [
    {
        "mode": "NULLABLE",
        "name": "CurrencyRateID",
        "type": "INTEGER"
    },
    {
        "mode": "NULLABLE",
        "name": "CurrencyRateDate",
        "type": "TIMESTAMP"
    },
    {
        "mode": "NULLABLE",
        "name": "FromCurrencyCode",
        "type": "STRING"
    },
    {
        "mode": "NULLABLE",
        "name": "ToCurrencyCode",
        "type": "STRING"
    },
    {
        "mode": "NULLABLE",
        "name": "AverageRate",
        "type": "FLOAT"
    },
    {
        "mode": "NULLABLE",
        "name": "EndOfDayRate",
        "type": "FLOAT"
    },
    {
        "mode": "NULLABLE",
        "name": "ModifiedDate",
        "type": "TIMESTAMP"
    }
    ]
EOF
}

resource "google_bigquery_table" "Sales_Customer" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_Customer"
  deletion_protection = false
  schema              = <<EOF
    [
    {
        "mode" : "NULLABLE",
        "name" : "CustomerID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "PersonID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "StoreID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "TerritoryID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "AccountNumber",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "rowguid",
        "type" : "STRING"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ModifiedDate",
        "type" : "TIMESTAMP"
    }
    ]
EOF
}

resource "google_bigquery_table" "Sales_PersonCreditCard" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_PersonCreditCard"
  deletion_protection = false
  schema              = <<EOF
    [
    {
        "mode" : "NULLABLE",
        "name" : "BusinessEntityID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "CreditCardID",
        "type" : "INTEGER"
    },
    {
        "mode" : "NULLABLE",
        "name" : "ModifiedDate",
        "type" : "TIMESTAMP"
    }
    ]
EOF
}

resource "google_bigquery_table" "Sales_SalesOrderDetail" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesOrderDetail"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "SalesOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesOrderDetailID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "CarrierTrackingNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "OrderQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "SpecialOfferID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "UnitPrice",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "UnitPriceDiscount",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "LineTotal",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_SalesOrderHeader" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesOrderHeader"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "SalesOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "RevisionNumber",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "OrderDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "DueDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ShipDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "Status",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "OnlineOrderFlag",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesOrderNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "PurchaseOrderNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "AccountNumber",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "CustomerID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesPersonID",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "TerritoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "BillToAddressID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ShipToAddressID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ShipMethodID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "CreditCardID",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "CreditCardApprovalCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "CurrencyRateID",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "SubTotal",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "TaxAmt",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Freight",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "TotalDue",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Comment",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_SalesOrderHeaderSalesReason" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesOrderHeaderSalesReason"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "SalesOrderID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesReasonID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_SalesPerson" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesPerson"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "TerritoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesQuota",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Bonus",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "CommissionPct",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesYTD",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesLastYear",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_SalesPersonQuotaHistory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesPersonQuotaHistory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "QuotaDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesQuota",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
  EOF
}

resource "google_bigquery_table" "Sales_SalesReason" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesReason"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "SalesReasonID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ReasonType",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
  EOF
}

resource "google_bigquery_table" "Sales_SalesTaxRate" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesTaxRate"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "SalesTaxRateID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StateProvinceID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "TaxType",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "TaxRate",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_SalesTerritory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesTerritory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "TerritoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "CountryRegionCode",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Group",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesYTD",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesLastYear",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "CostYTD",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "CostLastYear",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
  EOF
}

resource "google_bigquery_table" "Sales_SalesTerritoryHistory" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SalesTerritoryHistory"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "TerritoryID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "StartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "EndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
  EOF
}

resource "google_bigquery_table" "Sales_ShoppingCartItem" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_ShoppingCartItem"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "ShoppingCartItemID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ShoppingCartID",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Quantity",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "DateCreated",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_SpecialOffer" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SpecialOffer"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "SpecialOfferID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Description",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "DiscountPct",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "Type",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "Category",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "StartDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "EndDate",
    "type": "TIMESTAMP"
  },
  {
    "mode": "NULLABLE",
    "name": "MinQty",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "MaxQty",
    "type": "FLOAT"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_SpecialOfferProduct" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_SpecialOfferProduct"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "SpecialOfferID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "ProductID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
EOF
}

resource "google_bigquery_table" "Sales_Store" {
  dataset_id          = google_bigquery_dataset.dataset.dataset_id
  table_id            = "Sales_Store"
  deletion_protection = false
  schema              = <<EOF
  [
  {
    "mode": "NULLABLE",
    "name": "BusinessEntityID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Name",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "SalesPersonID",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "Demographics",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "rowguid",
    "type": "STRING"
  },
  {
    "mode": "NULLABLE",
    "name": "ModifiedDate",
    "type": "TIMESTAMP"
  }
]
  EOF
}