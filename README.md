# CSB Winter Term 2023/2024

Welcome to my CSB-Project! 
This benchmarking project aims at answering the following research question:

> "How does scaling-out affect throughput in RabbitMQ cluster?"

This file contains information on the structure of my project and on how to run the code.

## Structure

My benchmark setup consists of 3 kinds of nodes: producer, broker, and consumer.
The producer nodes generate and send messages, the broker nodes run the SUT, and the consumer nodes collect the data relevant for throughput.
In addition to the SUT itself, the `monitoring` code also runs on the broker nodes.
It collects CPU, memory, and network data on each broker node during the experiment.

The `infrastructure` directory contains the necessary scripts to create the setup.

The `lib` directory contains code shared between the producer, broker (=> `monitoring`), and consumer.

The `analysis` directory (surprisingly!) contains code for the data analysis.

## Running the code

### Creating the infrastructure

The scripts to create the infrastructure are in the `infrastructure` directory.

#### Creating the VMs

All VMs and networking are created using terraform.
To create the infrastructure, just run

```shell
terraform apply --auto-approve
```

in the terraform subdirectory.
This will also execute the startup scripts for the producer, consumer, and broker nodes.
These install all necessary (and some nice-to-have) packages and download + build the code.
*Note*: The variable `broker_count` in the file `./terraform/locals.tf` specifies the number of broker nodes for this experiment run.
This needs to be manually changed to 3, 5, or 7 (at least for my experiment).
Alternatively, just leave it at 7.

#### Creating the cluster

To create the cluster, run the following commands.

```shell
N_BROKER_NODES=3 # or 5 or 7, depending on the current run
./setup/setup_cluster.sh "$N_BROKER_NODES"
```

### Starting the experiment 

> **Note to self**: The `terraform apply --auto-approve` command returns before all startup scripts have finished.
> I.e., wait a bit before starting the monitoring etc., as the code will have to be built first.

You can start the experiment by running the following commands.

```shell
N_PROD=2
N_CONS=3
N_BROK=7
./setup/start_experiment.sh "$N_PROD" "$N_CONS" "$N_BROK"
```

### Monitoring during the experiment

To check the experiment status during a run, use the RabbitMQ management UI under `http:<broker-ip>:15672` or `ssh` into the VMs and use `htop`.
To log into the management console, use the credentials "jeff": "jeff".

### Stopping the experiment

The experiment stops automatically after the duration specified in `cloud-service-benchmarking/config.json`.
Afterwards, the producers stop sending the regular workload and notify the consumers that the experiment is over.
After the consumers have emptied the queue, they write the measurements remaining in their buffers to disk, which could take a few moments.

Consequently, only the data collection on the broker nodes needs to be stopped manually.
To do so, just the following code snippet.

```shell
for (( i = 0 ; i < "$N_BROK" ; i++ )); do
	gcloud compute ssh "broker-instance-$i" --command \
		'sudo kill -s 2 "$(pgrep main | head -n 1)"'
done
```

### Analysis

The `analysis` directory contains two python scripts.
First, run this

```shell
python3 create_database.py
```

to create a database from the experiment data. 
Make sure to move the data into the correct directory as specified in the `create_database.py` file.

Lastly, run the analysis script to create the plots.

```shell
python3 analysis.py
```
