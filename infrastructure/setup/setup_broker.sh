#!/bin/sh

#
# gcloud parameters
#

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
    "sudo apt-get install -y rabbitmq-server ranger network-manager net-tools"
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
# transfer monitoring code
#

# relative paths & required files/directories:
# - benchmark
#   - monitoring/
#   - lib/
#   - terraform/
#       - setup/ <= you are here

gcloud compute scp --recurse "./../../monitoring" "$user"@"$instance":~
gcloud compute scp --recurse "./../../lib" "$user"@"$instance":~

#
# build it
#

commands=(
    "cd ~/monitoring; /usr/local/go/bin/go build cmd/main.go"
)

for command in "${commands[@]}"; do
    execute_remote "$instance" "$zone" "$command"
done

#
# setup rabbitmq
#

commands=(
    # erlang cookie must be the same on all nodes for clustering
    # the content itself doesn't matter (just a string),
    # it just needs to be the same across all nodes
    # the hash should be `T01lC+isYECECAAyiG87Uw==`
    # to find out, run: (and adjust the instance name if necessary)
    # `sudo cat /var/log/rabbitmq/rabbit@producer-instance-0.log | grep cookie`
    "sudo rabbitmq-plugins enable rabbitmq_management"
    "sudo chmod 666 /var/lib/rabbitmq/.erlang.cookie"
    'echo "mynameisjeff" > /var/lib/rabbitmq/.erlang.cookie'
    "sudo chmod 600 /var/lib/rabbitmq/.erlang.cookie"
    "sudo systemctl restart rabbitmq-server"
    # create a new user (guest only works on localhost)
    # this is also used by the consumers + producers
    "sudo rabbitmqctl add_user 'jeff' 'jeff'"
    'sudo rabbitmqctl set_permissions -p "/" "jeff" ".*" ".*" ".*"'
    "sudo rabbitmqctl set_user_tags jeff administrator"
    # required for cluster formation
    'for (( i = 0 ; i < 7 ; i++ )); do echo "10.0.0.$((i+2)) producer-instance-$i" | sudo tee -a /etc/hosts; done'
)

for command in "${commands[@]}"; do
    execute_remote "$instance" "$zone" "$command"
done
