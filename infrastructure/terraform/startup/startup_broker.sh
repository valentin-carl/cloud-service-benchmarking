exec > >(tee -a /var/log/startup.log) 2>&1

# start + check internet connection through cloud NAT
echo "startup script executed"
env
curl neverssl.com

# install packages
sudo apt-get update
sudo apt-get install -y zsh ranger htop network-manager net-tools git vim curl iputils-ping

# install + setup rabbitMQ
sudo apt-get install -y rabbitmq-server
sudo rabbitmq-plugins enable rabbitmq_management
sudo chmod 666 /var/lib/rabbitmq/.erlang.cookie
echo "mynameisjeff" > /var/lib/rabbitmq/.erlang.cookie
sudo chmod 600 /var/lib/rabbitmq/.erlang.cookie
sudo systemctl restart rabbitmq-server
sudo rabbitmqctl set_vm_memory_high_watermark 0.7
sudo rabbitmqctl add_user "jeff" "jeff"
sudo rabbitmqctl set_permissions -p "/" "jeff" ".*" ".*" ".*"
sudo rabbitmqctl set_user_tags jeff administrator
for (( i = 0 ; i < 7 ; i++ )); do
    echo "10.0.0.$((i+2)) broker-instance-$i" | sudo tee -a /etc/hosts
done

# install golang
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
echo 'Defaults        secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/usr/local/go/bin"' | sudo tee /etc/sudoers.d/spath

# get + build monitoring code
mkdir benchmark
cd benchmark
git clone https://github.com/valentin-carl/cloud-service-benchmarking.git
cd cloud-service-benchmarking/monitoring
sudo go build ./cmd/main.go

# end startup
ls -lah
echo done
