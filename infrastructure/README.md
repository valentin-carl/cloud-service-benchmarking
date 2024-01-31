# Setup, Running the Experiment, and Data Collection

> *Note*: All commands in this README assume that `cloud-service-benchmarking/infrastructure` is the current working directory.

## 0. Creating the infrastructure

All VMs and networking is created using terraform.
Do create the infrastructure, just run 

```shell
terraform apply --auto-approve
```

in the terraform subdirectory.
This will also execute the startup scripts for the producer, consumer, and broker nodes.
These install all necessary (and some nice-to-have) packages and download + build the code.
*Note*: The variable `broker_count` in the file `./terraform/locals.tf` specifies the number of broker nodes for this experiment run.
This needs to be manually changed to 3, 5, or 7 (at least for my experiment).
In theory, it could be left at seven and the cluster size determined in the next step, but that would result in VMs running unnecessarily for an hour.

## 1. Creating the cluster

To create the cluster, run the following commands.

```shell
N_BROKER_NODES=3 # or 5 or 7, depending on the current run
./setup/setup_cluster.sh "$N_BROKER_NODES"
```

## 2. Running the experiment

### A. Starting the experiment

TODO 

### B. Checking on the nodes during the experiment 

TODO

### C. Stopping the experiment

The experiment stop automatically after the duration specified in `cloud-service-benchmarking/config.json`.
Afterwards, the producers stop sending the regular workload and notify the consumers that the experiment is over.
After the consumers have emptied the queue, they write the measurements remaining in their buffers to disk, which could take a few moments.

TODO how to stop the broker data collection ????

## 3. Collecting the data

To collect the CPU, memory, and network measurements from each broker node, run the following commands for each node.

```shell
LOCAL_DIR = "..." # this is where the data will be downloaded to
gcloud compute ccp broker-instance-0:/benchmark/cloud-service-benchmarking/monitoring/data "$LOCAL_DIR" # adjust the instance name accordingly
```

TODO collect data from consumers => test if fast enough through scp or if google storage bucket is necessary

## 4. Destroying the infrastructure

To destroy the infrastructure after an experiment run, use terraform again.

```shell
terraform destroy --auto-approve
```

----

### Broker instances

#### Start monitoring at all brokers

> **Note to self**: The `terraform apply -auto-approve` command returns before all startup scripts have finished.
> I.e., wait a bit before starting the monitoring etc., as the code will have to be built first.

```shell
for (( i = 0 ; i < "$N_BROK" ; i++ )); do
	gcloud compute ssh "broker-instance-$i" --command \
		"export NODEID=$i; sudo -E bash -c '/benchmark/cloud-service-benchmarking/monitoring/main &>/dev/null &'"
done
```

#### Stop monitoring at all brokers

The following code snipped sends interrupt signals to the monitoring processes.
These, in turn, listen for SIGINT to stop gracefully.
After the signal has been received, the remaining CPU, memory, and networking data is written to disk and the program is done.

```shell
for (( i = 0 ; i < "$N_BROK" ; i++ )); do
	gcloud compute ssh "broker-instance-$i" --command \
		'sudo kill -s 2 "$(pgrep main | head -n 1)"'
done
```

#### Getting the data

```shell
# todo
```
