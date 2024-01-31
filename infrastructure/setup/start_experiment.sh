# todo idea for nodeid stuff: when starting, do "export \$NODEID=0; ./main"
# at least this works:
# `gcloud compute ssh producer-instance-0 --zone=europe-west10-a --command='export A=1; env'`
# so the programs should be able to also see that environment variable??

# todo quorum queue configure replica count
# todo rabbitmq configure memory stuff 40% -> 70%


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
    gcloud compute ssh "consumer-instance-$i" --command "cd /benchmark/cloud-service-benchmarking/consumer; export NODEID=$i; sudo -E \$NODEID bash -c './main'" & # TODO test
done

# start monitoring at broker nodes
for (( i = 0 ; i < "$N_BROK" ; i++ )); do
    gcloud compute ssh "broker-instance-$i" --command "cd /benchmark/cloud-service-benchmarking/monitoring; export NODEID=$i; sudo -E \$NODEID bash -c './main'" & # TODO test
done

# start producers
for (( i = 0 ; i < "$N_PROD" ; i++ )); do
    gcloud compute ssh "producer-instance-$i" --command "cd /benchmark/cloud-service-benchmarking/producer; sudo ./main'" & # TODO test
done

# TODO check what happens if this script runs through? does that the processes running on the VMs? In that case, try moving the &s into the remote commands
# TODO redirect the output of these programs to /var/log/... using tee?









for (( i = 0 ; i < "$N_BROK" ; i++ )); do
    gcloud compute ssh "broker-instance-$i" --command \
        "cd /benchmark/cloud-service-benchmarking/monitoring; export NODEID=$i; sudo -E bash -c './main'" &
done










