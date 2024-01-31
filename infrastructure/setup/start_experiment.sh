#!/bin/sh

N_PROD=$1
N_CONS=$2
N_BROK=$3

if [ -z "$1" ]; then
    echo "missing number of producer nodes"
    exit
fi

if [ -z "$2" ]; then
    echo "missing number of consumer nodes"
    exit
fi

if [ -z "$3" ]; then
    echo "missing number of broker nodes"
    exit
fi

# start consumers
for (( i = 0 ; i < "$N_CONS" ; i++ )); do
    gcloud compute ssh "consumer-instance-$i" --command \
		"export NODEID=$i; sudo -E bash -c '/benchmark/cloud-service-benchmarking/consumer/main &>/dev/null &'"
done

# start monitoring at broker nodes
for (( i = 0 ; i < "$N_BROK" ; i++ )); do
	gcloud compute ssh "broker-instance-$i" --command \
		"export NODEID=$i; sudo -E bash -c '/benchmark/cloud-service-benchmarking/monitoring/main &>/dev/null &'"
done

# start producers
for (( i = 0 ; i < "$N_PROD" ; i++ )); do
    gcloud compute ssh "producer-instance-$i" --command \
        "sudo -E bash -c '/benchmark/cloud-service-benchmarking/producer/main &>/dev/null &'"
done
