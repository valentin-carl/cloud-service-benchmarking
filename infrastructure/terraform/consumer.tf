resource "google_compute_instance_template" "template-consumer" {
  name         = "template-consumer"
  machine_type = "e2-highmem-2"

  disk {
    source_image = "ubuntu-2204-jammy-v20240119a"
    auto_delete  = true
    disk_size_gb = 100
    boot         = true
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    access_config {}
  }
}

resource "google_compute_instance_from_template" "consumer" {
  count = local.consumer_count

  name = "consumer-instance-${count.index}"
  zone = "europe-west10-a"

  source_instance_template = google_compute_instance_template.template-consumer.self_link

  metadata_startup_script = file("./startup/startup_consumer.sh")

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.name
    network_ip = "10.0.2.${count.index}"
  }
}

# to start the consumer, run:
# `gcloud compute ssh "$instance" --zone="zone" --command="export $NODEID=3; cd ~/consumer; ./main"`
# todo check that the nodeid stuff works
