#!/bin/sh

N_NODES="$1"
if [ -z "$N_NODES" ]; then
    echo "need number of broker nodes to execute script"
    exit
fi

execute_remote () {
    local cmd="gcloud compute ssh $1 --zone='$2' --command='$3'"
    echo "executing remotely: \"$cmd\""
    echo "$(eval "$cmd")"
}

cluster() {
    commands=(
        "sudo rabbitmqctl stop_app"
        "sudo rabbitmqctl join_cluster rabbit@broker-instance-0"
        "sudo rabbitmqctl start_app"
    )
    for command in "${commands[@]}"; do
        execute_remote "$1" "europe-west10-a" "$command"
    done
}

# start at 1 because we join the cluster of node-0 and don't need to call this for that node
for (( i = 1 ; i < "$N_NODES" ; i++ )); do
    cluster "broker-instance-$i"
done
