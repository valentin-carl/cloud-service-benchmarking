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

# download monitoring data from brokers
for (( i = 0 ; i < "$N_BROK" ; i++ )); do
    gcloud compute scp --recurse "broker-instance-$i":~/ ./broker/
done

# download results from consumers
for (( i = 0 ; i < "$N_CONS" ; i++ )); do
    gcloud compute scp --recurse "consumer-instance-$i":/benchmark/cloud-service-benchmarking/consumer/out/ ./consumer/
    gcloud compute scp --recurse "consumer-instance-2":/benchmark/cloud-service-benchmarking/consumer/out/ ./consumer/
done

#
# LOGS
#

# producer
for (( i = 0 ; i < "$N_PROD" ; i++ )); do
    mkdir -p "./logs/producer-$i"
    gcloud compute scp "producer-instance-$i":/var/log/benchmark.log "./logs/producer-$i/benchmark.log"
    gcloud compute scp "producer-instance-$i":/var/log/startup.log "./logs/producer-$i/startup.log"
done

# broker
for (( i = 0 ; i < "$N_BROK" ; i++ )); do
    mkdir -p "./logs/broker-$i"
    gcloud compute scp "broker-instance-$i":/var/log/benchmark.log "./logs/broker-$i/benchmark.log"
    gcloud compute scp "broker-instance-$i":/var/log/startup.log "./logs/broker-$i/startup.log"
    gcloud compute scp --recurse "root@broker-instance-$i":/var/log/rabbitmq/ "./logs/broker-$i/"
done

# consumer
for (( i = 0 ; i < "$N_CONS" ; i++ )); do
    mkdir -p "./logs/consumer-$i"
    gcloud compute scp "consumer-instance-$i":/var/log/benchmark.log "./logs/consumer-$i/benchmark.log"
    gcloud compute scp "consumer-instance-$i":/var/log/startup.log "./logs/consumer-$i/startup.log"
done
