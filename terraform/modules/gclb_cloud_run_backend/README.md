# abcxyz Cloud Run Module

This module provides the default Global HTTPS Load Balancer for use with a Cloud
Run service as the backend for abcxyz projects.

## Updating Certificates

Update the `domains` variable by adding any additional domains you want to
redirect to your load balancer.

1. Make sure to provide the load balancer's IP address (provided as an output of
   this module) to your domain record managed by DNS. Failure to do so can cause
   a prolonged outage.
2. Provide your entries into the `domains` list in order to provision a cert,
   forwarding rule, and target proxy. The order of the `domains` list matters,
   if you change the order of the existing domains this will cause a new
   certificate to be created. i.e. `[A, B]` to `[B, A]` will trigger a new cert
   to be created.
3. New resources will be provisioned based on the latest `domains` list. The
   prevous certificate that managed the previous list of domains will be
   removed. This will cause a temporary outage until the new cert is provisioned
   which takes at most 1 hour.
4. Apply the terraform using this module, monitor your new certificate in the
   console and wait for it to say "ACTIVE" next to the status of each domain in
   the list. Find this on the "Certificate Manager" page or by viewing your load
   balancer configuration in the console.
5. Once the status is confirmed your new certificate is up and running.

<!-- BEGIN_TF_DOCS -->
## Examples

```terraform
module "gclb_cloud_run_backend" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=SHA_OR_TAG"

  project_id = "my-project-id"

  name             = "project-name"
  run_service_name = "service-name"
  domains          = ["project.company.domain.com"]
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_domains"></a> [domains](#input\_domains) | Domain names to use for the HTTPS Global Load Balancer for the Cloud Run service (e.g. ["my-project.e2e.tycho.joonix.net"]). | `list(string)` | n/a | yes |
| <a name="input_iap_config"></a> [iap\_config](#input\_iap\_config) | Identity-Aware Proxy configuration for the load balancer. | <pre>object({<br>    enable               = bool<br>    oauth2_client_id     = string<br>    oauth2_client_secret = string<br>  })</pre> | <pre>{<br>  "enable": false,<br>  "oauth2_client_id": "",<br>  "oauth2_client_secret": ""<br>}</pre> | no |
| <a name="input_name"></a> [name](#input\_name) | The name of this project. | `string` | n/a | yes |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The Google Cloud project ID. | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | The default Google Cloud region to deploy resources in (defaults to 'us-central1'). | `string` | `"us-central1"` | no |
| <a name="input_run_service_name"></a> [run\_service\_name](#input\_run\_service\_name) | The name of the Cloud Run service to the compute backend serverless network endpoint group. | `string` | n/a | yes |
| <a name="input_security_policy"></a> [security\_policy](#input\_security\_policy) | Cloud Armor security policy for the load balancer. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_external_ip_address"></a> [external\_ip\_address](#output\_external\_ip\_address) | The external IPv4 assigned to the global fowarding rule. |
| <a name="output_external_ip_name"></a> [external\_ip\_name](#output\_external\_ip\_name) | The external IPv4 name assigned to the global fowarding rule. |

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 4.45 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | >= 4.45 |
| <a name="provider_random"></a> [random](#provider\_random) | n/a |

## Resources

| Name | Type |
|------|------|
| [google_compute_backend_service.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_backend_service) | resource |
| [google_compute_global_address.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_global_address) | resource |
| [google_compute_global_forwarding_rule.http](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_global_forwarding_rule) | resource |
| [google_compute_global_forwarding_rule.https](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_global_forwarding_rule) | resource |
| [google_compute_managed_ssl_certificate.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_managed_ssl_certificate) | resource |
| [google_compute_region_network_endpoint_group.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_region_network_endpoint_group) | resource |
| [google_compute_target_http_proxy.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_target_http_proxy) | resource |
| [google_compute_target_https_proxy.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_target_https_proxy) | resource |
| [google_compute_url_map.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_url_map) | resource |
| [google_compute_url_map.https_redirect](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_url_map) | resource |
| [google_project_service.services](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/project_service) | resource |
| [random_id.cert](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |
| [random_id.default](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |

## Modules

No modules.
<!-- END_TF_DOCS -->
