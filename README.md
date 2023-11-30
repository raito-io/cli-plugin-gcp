<h1 align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/raito-io/raito-io.github.io/raw/master/assets/images/logo-vertical-dark%402x.png">
    <img height="250px" src="https://github.com/raito-io/raito-io.github.io/raw/master/assets/images/logo-vertical%402x.png">
  </picture>
</h1>

<h4 align="center">
  Google Cloud Platform plugin for the Raito CLI
</h4>

<p align="center">
    <a href="/LICENSE.md" target="_blank"><img src="https://img.shields.io/badge/license-Apache%202-brightgreen.svg" alt="Software License" /></a>
    <a href="https://github.com/raito-io/cli-plugin-gcp/actions/workflows/build.yml" target="_blank"><img src="https://img.shields.io/github/actions/workflow/status/raito-io/cli-plugin-gcp/build.yml?branch=main" alt="Build status"/></a>
    <a href="https://codecov.io/gh/raito-io/cli-plugin-gcp" target="_blank"><img src="https://img.shields.io/codecov/c/github/raito-io/cli-plugin-gcp" alt="Code Coverage" /></a>
</p>

<hr/>

# Raito CLI Plugin - Google Cloud Platform


**Note: This repository is still in an early stage of development.
At this point, no contributions are accepted to the project yet.**

Two plugins are build from the source code in this repository to support two GCP features:
1. Google Cloud Platform: GCP Plugin
2. BigQuery: BQ Plugin

## Raito CLI Plugin - Google Cloud Platform

This Raito CLI plugin implements the integration with Google Cloud Platform. It can
 - Synchronize the users and groups in GSuite
 - Synchronize the users, groups and service accounts bound in GCP Projects and Folders
 - Synchronize the GCP organizational structure (folders, projects) to a data source in Raito Cloud along with the access controls in place.
 - Synchronize the access controls from Raito Cloud (or from a file in case of the [access-as-code flow](http://docs.raito.io/docs/guide/access)) into IAM permissions


### Prerequisites
To use this plugin, you will need

1. The Raito CLI to be correctly installed. You can check out our [documentation](http://docs.raito.io/docs/cli/installation) for help on this.
2. A Raito Cloud account to synchronize your GCP organization with. If you don't have this yet, visit our webpage at (https://raito.io) and request a trial account.
3. A service account to a GCP project. This service account should be able to access all folders/projects IAM policies you want to sync
4. If you wish to sync identities from GSuite, the SA needs domain-wide-delegation set up in GSuite Admin Console

A full example on how to start using Raito Cloud with Snowflake can be found as a [guide in our documentation](http://docs.raito.io/docs/guide/cloud).

### Usage
To use the plugin, add the following snippet to your Raito CLI configuration file (`raito.yml`, by default) under the `targets` section:

```json
  - name: gcp1
    connector-name: raito-io/cli-plugin-gcp/gcp
    data-source-id: <<GCP datasource ID>>   
    identity-store-id: <<GCP identitystore ID>>
    
    gcp-serviceaccount-json-location: <<location_to_sa_json>>
    gcp-organization-id: <<gcp_org_id>>

    gsuite-identity-store-sync: true/false
    gsuite-customer-id: <<GSuite Customer ID>>
    gsuite-impersonate-subject: <<GSuite impersonation subject>>
    

```

Next, replace the values of the indicated fields with your specific values:
- `<<GCP datasource ID>>`: the ID of the Data source you created in the Raito Cloud UI.
- `<<GCP identitystore ID>>`: the ID of the Identity Store you created in the Raito Cloud UI.
- `<<location_to_sa_json>>`: location of the JSON file containing the GCP serviceaccount credentials to use for synchronization. If not set, GOOGLE_APPLICATION_CREDENTIALS env var is used instead.
- `<<gcp_org_id>>`: The ID of the GCP organization which you retreive form the Google Cloud Platform Console or GCP CLI.
- `<<gsuite-identity-store-sync>>`: if set to true, users and groups will be synced from the GSuite Workspace (requires additional access rights). If false, only users/groups part of the IAM policies in the project are synced.
- `<<GSuite Customer ID>>`: (required when gsuite-identity-store-sync) The Customer ID of the GSuite Workspace (https://support.google.com/a/answer/10070793?hl=en)
- `<<GSuite impersonation subject>>`: (required when gsuite-identity-store-sync) The username of the GSuite administrator your service account will impersonate to contact the GSuite Directory API


You will also need to configure the Raito CLI further to connect to your Raito Cloud account, if that's not set up yet.
A full guide on how to configure the Raito CLI can be found on (http://docs.raito.io/docs/cli/configuration).

### Trying it out

As a first step, you can check if the CLI finds this plugin correctly. In a command-line terminal, execute the following command:
```bash
$> raito info raito-io/cli-plugin-gcp/gcp
```

This will download the latest version of the plugin (if you don't have it yet) and output the name and version of the plugin, together with all the plugin-specific parameters to configure it.

When you are ready to try out the synchronization for the first time, execute:
```bash
$> raito run
```
This will take the configuration from the `raito.yml` file (in the current working directory) and start a single synchronization.

Note: if you have multiple targets configured in your configuration file, you can run only this target by adding `--only-targets gcp1` at the end of the command.

## Raito CLI Plugin - BigQuery

**Note: This repository is still in an early stage of development.
At this point, no contributions are accepted to the project yet.**

This Raito CLI plugin implements the integration with Google Cloud BigQuery. It can
- Synchronize the users in a GCP Project/GSuite workspace to an identity store in Raito Cloud.
- Synchronize the BigQuery meta data (data structure, known permissions, ...) to a data source in Raito Cloud.
- Synchronize the access control from Raito Cloud (or from a file in case of the [access-as-code flow](http://docs.raito.io/docs/guide/access)) into IAM/BigQuery permissions
- Synchronize the data usage information to Raito Cloud.


### Prerequisites
To use this plugin, you will need

1. The Raito CLI to be correctly installed. You can check out our [documentation](http://docs.raito.io/docs/cli/installation) for help on this.
2. A Raito Cloud account to synchronize your Snowflake account with. If you don't have this yet, visit our webpage at (https://raito.io) and request a trial account.
3. A service account to a GCP project
4. If you wish to sync identities from GSuite, the SA needs domain-wide-delegation set up in GSuite Admin Console

A full example on how to start using Raito Cloud with Snowflake can be found as a [guide in our documentation](http://docs.raito.io/docs/guide/cloud).

### Usage
To use the plugin, add the following snippet to your Raito CLI configuration file (`raito.yml`, by default) under the `targets` section:

```json
  - name: bigquery1
    connector-name: raito-io/cli-plugin-gcp/bigquery
    data-source-id: <<BQ datasource ID>>   
    identity-store-id: <<BQ identitystore ID>>
    gcp-project-id: <<Google Cloud Platform Project ID>>

    gcp-serviceaccount-json-location: <<location_to_sa_json>>
    
    gsuite-identity-store-sync: true/false
    gsuite-customer-id: <<GSuite Customer ID>>
    gsuite-impersonate-subject: <<GSuite impersonation subject>>
    

```

Next, replace the values of the indicated fields with your specific values:
- `<<BQ datasource ID>>`: the ID of the Data source you created in the Raito Cloud UI.
- `<<BQ identitystore ID>>`: the ID of the Identity Store you created in the Raito Cloud UI.
- `>>location_to_sa_json>>`: location of the JSON file containing the GCP serviceaccount credentials to use for synchronization. If not set, GOOGLE_APPLICATION_CREDENTIALS env var is used instead.
- `gsuite-identity-store-sync`: if set to true, users and groups will be synced from the GSuite Workspace (requires additional access rights). If false, only users/groups part of the IAM policies in the project are synced.
- `<<GSuite Customer ID>>`: (required when gsuite-identity-store-sync) The Customer ID of the GSuite Workspace (https://support.google.com/a/answer/10070793?hl=en)
- `<<GSuite impersonation subject>>`: (required when gsuite-identity-store-sync) The username of the GSuite administrator your service account will impersonate to contact the GSuite Directory API


You will also need to configure the Raito CLI further to connect to your Raito Cloud account, if that's not set up yet.
A full guide on how to configure the Raito CLI can be found on (http://docs.raito.io/docs/cli/configuration).

### Trying it out

As a first step, you can check if the CLI finds this plugin correctly. In a command-line terminal, execute the following command:
```bash
$> raito info raito-io/cli-plugin-bigquery
```

This will download the latest version of the plugin (if you don't have it yet) and output the name and version of the plugin, together with all the plugin-specific parameters to configure it.

When you are ready to try out the synchronization for the first time, execute:
```bash
$> raito run
```
This will take the configuration from the `raito.yml` file (in the current working directory) and start a single synchronization.

Note: if you have multiple targets configured in your configuration file, you can run only this target by adding `--only-targets bigquery1` at the end of the command.
