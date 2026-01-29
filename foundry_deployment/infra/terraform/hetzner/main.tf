terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.48.0"
    }
  }
}

variable "hetzner_api_key" {
  type      = string
  sensitive = true
  default   = "invalid"
}

locals {
  location         = "hil"
  zone             = "us-west"
  network_ip_range = "10.0.0.0/8"
  server_ip_range  = "10.10.0.0/16"
}

provider "hcloud" {
  token = var.hetzner_api_key
}

resource "hcloud_ssh_key" "foundry_ssh_key" {
  name = "foundry_ssh_key"
  public_key = file("${path.root}/../escrow.pub")
}

resource "hcloud_network" "network" {
  name     = "foundry-network"
  ip_range = local.network_ip_range
}

resource "hcloud_load_balancer" "load_balancer" {
  name               = "foundry-balancer"
  load_balancer_type = "lb11"
  location           = local.location
}

resource "hcloud_network_subnet" "server_subnet" {
  ip_range     = local.server_ip_range
  network_id   = hcloud_network.network.id
  network_zone = local.zone
  type         = "cloud"
}

resource "hcloud_load_balancer_network" "balancer_network" {
  subnet_id               = hcloud_network_subnet.server_subnet.id
  load_balancer_id        = hcloud_load_balancer.load_balancer.id
  enable_public_interface = true
}

resource "hcloud_server" "server" {
  name        = "foundry-server"
  image       = "debian-12"
  server_type = "cpx21"
  location    = local.location
  ssh_keys = [hcloud_ssh_key.foundry_ssh_key.id]
  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }
}

resource "hcloud_server_network" "server_network" {
  server_id = hcloud_server.server.id
  subnet_id = hcloud_network_subnet.server_subnet.id
}

resource "hcloud_load_balancer_target" "balancer_target" {
  type             = "server"
  load_balancer_id = hcloud_load_balancer.load_balancer.id
  server_id        = hcloud_server.server.id
  use_private_ip   = true
}

# module "managed_cert" {
#   source      = "./managed_cert"
#   balancer_id = hcloud_load_balancer.load_balancer.id
# }