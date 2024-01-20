#!/bin/sh

# gcloud parameters
instance="$1"
if [ -z "$1" ]; then
    echo "need instance name to execute script"
    exit
fi
if [ -z "$2" ]; then
    zone="europe-west10-a"
else
    zone="$2"
fi
if [ -z "$3" ]; then
    user="valentincarl"
else
    user="$3"
fi
echo "input parameters"
echo "> instance name: $instance"
echo "> zone:\t\t $zone"
echo "> user:\t\t $user"

# execute a command remotely on a compute instance
# inputs:
# - $1: instance name
# - $2: zone
# - $3: command
execute_remote () {
    local gcmd="gcloud compute ssh $1 --zone='$2' --command='$3'"
    echo "executing remotely: \"$gcmd\""
    echo "$(eval "$gcmd")"
}

#
# install packages
#

commands=(
    "sudo apt-get update"
    "sudo apt-get install -y fish ranger network-manager net-tools"
    #"sudo apt-get install -y bash htop vim curl iputils-ping" # these should already be installed
)

for command in "${commands[@]}"; do
    execute_remote "$instance" "$zone" "$command"
done

#
# install golang
#

commands=(
    "wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz"
    "sudo rm -rf /usr/local/go"
    "sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz"
    'echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc'
    "/usr/local/go/bin/go version"
)

for command in "${commands[@]}"; do
    execute_remote "$instance" "$zone" "$command"
done

#
# send producer code to VM
#

# relative paths & required files/directories:
# - benchmark
#   - producer/
#   - lib/
#   - terraform/
#       - setup/ <= you are here
#   - config.json

gcloud compute scp --recurse "./../../producer" "$user"@"$instance":~
gcloud compute scp --recurse "./../../lib" "$user"@"$instance":~
gcloud compute scp "./../../config.json" "$user"@"$instance":~

#
# build the producer
#

commands=(
    "cd ~/producer; /usr/local/go/bin/go build cmd/main.go"
)

for command in "${commands[@]}"; do
    execute_remote "$instance" "$zone" "$command"
done

# to start the producer, run:
# `gcloud compute ssh "$instance" --zone="zone" --command="cd ~/producer; ./main"`
