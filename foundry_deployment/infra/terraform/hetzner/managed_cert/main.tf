terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.48.0"
    }
  }
}

resource "hcloud_managed_certificate" "managed_cert" {
  name = "noodles_managed_cert"
  domain_names = ["*.noodles.quest", "noodles.quest"]
}

resource "hcloud_load_balancer_service" "balancer_service" {
  load_balancer_id = var.balancer_id
  protocol         = "https"
  listen_port      = 443
  destination_port = 30000

  http {
    certificates = [hcloud_managed_certificate.managed_cert.id]
  }
}