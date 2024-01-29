exec > >(tee -a /var/log/startup.log) 2>&1

echo "startup script executed"
env
curl neverssl.com # ensure internet connection through nat works

sudo apt-get update
sudo apt-get install -y zsh ranger htop network-manager net-tools git vim curl iputils-ping

wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version

echo 'Defaults        secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/usr/local/go/bin"' | sudo tee /etc/sudoers.d/spath

mkdir benchmark
cd benchmark

git clone https://github.com/valentin-carl/cloud-service-benchmarking.git
cd cloud-service-benchmarking/consumer
sudo go build ./cmd/main.go

ls -lah
echo done
