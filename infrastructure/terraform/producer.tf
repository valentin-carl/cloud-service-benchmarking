resource "google_compute_instance_template" "template-producer" {
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

resource "google_compute_instance_from_template" "producer" {
  count = local.producer_count

  name = "producer-instance-${count.index}"
  zone = "europe-west10-c"

  source_instance_template = google_compute_instance_template.template-producer.self_link_unique

  metadata_startup_script = file("./startup/startup_producer.sh")

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    network_ip = "10.0.1.${count.index}"
  }
}

# to start the producer, run:
# `gcloud compute ssh "$instance" --zone="zone" --command="cd ~/producer; ./main"`
