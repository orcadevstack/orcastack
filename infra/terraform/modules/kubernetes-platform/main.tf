terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }
}

variable "namespace" {
  type = string
}

variable "ingress_floating_ip" {
  type = string
}

variable "barbican_secret_id" {
  type = string
}

variable "ceph_storage_class_name" {
  type = string
}

variable "runner_service_account" {
  type = string
}

resource "kubernetes_namespace_v1" "platform" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/part-of" = "orcastack"
      "orcastack.io/governance"   = "enforced"
    }
  }
}

resource "kubernetes_service_account_v1" "runner" {
  metadata {
    name      = var.runner_service_account
    namespace = kubernetes_namespace_v1.platform.metadata[0].name
    annotations = {
      "orcastack.io/identity-mode" = "keystone-linked"
    }
  }
}

resource "kubernetes_storage_class_v1" "ceph" {
  metadata {
    name = var.ceph_storage_class_name
    annotations = {
      "orcastack.io/storage-tier" = "artifacts"
    }
  }

  storage_provisioner    = "csi.ceph.com"
  reclaim_policy         = "Retain"
  volume_binding_mode    = "WaitForFirstConsumer"
  allow_volume_expansion = true

  parameters = {
    clusterID = "orcastack-private-cloud"
    pool      = "orcastack-artifacts"
  }
}

resource "kubernetes_config_map_v1" "platform_bootstrap" {
  metadata {
    name      = "orcastack-platform-bootstrap"
    namespace = kubernetes_namespace_v1.platform.metadata[0].name
  }

  data = {
    INGRES_FIP         = var.ingress_floating_ip
    BARBICAN_SECRET_ID = var.barbican_secret_id
  }
}

output "namespace" {
  value = kubernetes_namespace_v1.platform.metadata[0].name
}