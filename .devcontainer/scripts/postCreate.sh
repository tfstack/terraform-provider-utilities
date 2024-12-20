# Include commands that you would like to execute after the container is created
uname -a

apt-get update -y
export DEBIAN_FRONTEND=noninteractive

# terraform
apt-get install -y apt-utils gnupg software-properties-common
curl -s https://apt.releases.hashicorp.com/gpg | gpg --dearmor > hashicorp.gpg
install -o root -g root -m 644 hashicorp.gpg /etc/apt/trusted.gpg.d/
apt-add-repository -y "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
apt-get update -y
apt-get install -y terraform
rm hashicorp.gpg

# install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.0

# install tfplugindocs
export GOBIN=$PWD/bin
export PATH=$GOBIN:$PATH
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
