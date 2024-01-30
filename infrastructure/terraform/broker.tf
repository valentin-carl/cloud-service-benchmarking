resource "google_compute_instance_template" "template-broker" {
  name         = "template-broker"
  machine_type = "e2-standard-2"

  disk {
    source_image = "ubuntu-2204-jammy-v20240119a"
    auto_delete  = true
    disk_size_gb = 25
    boot         = true
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    access_config {}
  }
}

resource "google_compute_instance_from_template" "broker" {
  count = local.broker_count

  name = "broker-instance-${count.index}"
  zone = "europe-west10-a"

  source_instance_template = google_compute_instance_template.template-broker.self_link

  metadata_startup_script = file("./startup/startup_broker.sh")

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    network_ip = "10.0.0.${count.index + 2}"
    access_config {}
  }
}
