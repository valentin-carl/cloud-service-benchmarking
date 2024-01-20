terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.12.0"
    }
  }
}

provider "google" {
  credentials = file("credentials.json")

  project = "prime-bonbon-407317"
  region  = "europe-west10"
  zone    = "europe-west10-a"
}

resource "google_compute_network" "benchnet" {
  name                    = "benchnet"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  name          = "subnet"
  region        = "europe-west10"
  network       = google_compute_network.benchnet.id
  ip_cidr_range = "10.0.0.0/22"
  // subnet:
  // producers: 10.0.0.<producer>
  // brokers:   10.0.1.<broker>
  // consumer:  10.0.2.<consumer>
  // remember that google reserves the first two (among others) addresses in every subnet
  // => https://cloud.google.com/vpc/docs/subnets#valid-ranges
}

resource "google_compute_firewall" "allow-ssh" {
  name    = "allow-ssh"
  network = google_compute_network.benchnet.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_firewall" "allow-ping" {
  name    = "allow-ping"
  network = google_compute_network.benchnet.name

  allow {
    protocol = "icmp"
  }

  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_firewall" "allow-rabbit" {
  name    = "allow-rabbit"
  network = google_compute_network.benchnet.name

  allow {
    protocol = "tcp"
    ports    = ["5672", "25672"]
  }

  // 5672 is used to send messages to and by the broker
  // 25672 is for cluster formation
  source_ranges = ["10.0.0.0/22"]
}

resource "google_compute_firewall" "allow-rabbit-management" {
  name    = "allow-rabbit-management"
  network = google_compute_network.benchnet.name

  allow {
    protocol = "tcp"
    // todo test if accessible from outside vpc
    // 15672 is used to access the management UI
    ports = ["15672"]
  }

  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_instance_template" "template-produce" {
  name         = "template-producer"
  machine_type = "e2-standard-2"

  disk {
    source_image = "ubuntu-2204-jammy-v20240119a"
    auto_delete  = true
    disk_size_gb = 10
    boot         = true
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    access_config {}
  }
}

resource "google_compute_instance_from_template" "producer-0" {
  name = "producer-instance-2"
  zone = "europe-west10-a"

  source_instance_template = google_compute_instance_template.template-produce.self_link_unique

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    network_ip = "10.0.0.2"
    access_config {}
  }
}

resource "google_compute_instance_from_template" "producer-1" {
  name = "producer-instance-3"
  zone = "europe-west10-a"

  source_instance_template = google_compute_instance_template.template-produce.self_link_unique

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    network_ip = "10.0.0.3"
    access_config {}
  }
}
