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
    ports    = [
      "4369",
      "5672",
      "6000-6500",
      "15672",
      "25672",
      "35672-35682"
    ]
  }

  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_router" "router" {
  name = "router"
  region = google_compute_subnetwork.subnet.region
  network = google_compute_network.benchnet.name

  bgp {
    asn = 64514
  }
}

resource "google_compute_router_nat" "nat" {
  name                   = "nat"
  region                 = google_compute_router.router.region
  nat_ip_allocate_option = "AUTO_ONLY"
  router                 = google_compute_router.router.name
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"

  # TODO
}