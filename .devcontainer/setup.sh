// setup.sh
# install curl, git, ...
apt-get update
apt-get install -y curl git jq

useradd -m user
su user

# install go
VERSION='1.23.2'
OS='linux'
ARCH='amd64'

curl -OL https://dl.google.com/go/go${VERSION}.${OS}-${ARCH}.tar.gz
tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz
rm go$VERSION.$OS-$ARCH.tar.gz

INSTALLED_GO_VERSION=$(go version)
echo "Go version ${INSTALLED_GO_VERSION} is installed"

# install gopls, dlv, hey
echo "Getting development tools"
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/rakyll/hey@latest

# vscode-go dependencies 
echo "Getting dependencies for the vscode-go plugin "
# via: https://github.com/microsoft/vscode-go/blob/master/.travis.yml
go install github.com/acroca/go-symbols
go install github.com/cweill/gotests@latest
go install github.com/davidrjenni/reftools/cmd/fillstruct@latest
go install github.com/haya14busa/goplay/cmd/goplay@latest
go install github.com/mdempsky/gocode@latest
go install github.com/ramya-rao-a/go-outline@latest
go install github.com/rogpeppe/godef@latest
go install github.com/sqs/goreturns@latest
go install github.com/uudashr/gopkgs/cmd/gopkgs@latest
go install github.com/zmb3/goinstalldoc@latest
go install golang.org/x/lint/golint@latest
go install golang.org/x/tools/cmd/gorename@latest