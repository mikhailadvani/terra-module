# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

test: install-deps clean fmt vet
	mkdir -p pkg/terramodule/testdata/output
	go test -v ./pkg/... ./cmd/... -coverprofile cover.out.tmp
	cat cover.out.tmp > cover.out && rm cover.out.tmp
	go tool cover -func cover.out

build: install-deps
	mkdir -p build
	GOOS=$(uname | awk '{print tolower($0)}')
	CGO_ENABLED=0 go build -o build/terra-module main.go

build-all: install-deps
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux go build -o build/terra-module-linux main.go
	CGO_ENABLED=0 GOOS=darwin go build -o build/terra-module-mac main.go

clean:
	rm -rf build
	rm -rf pkg/terramodule/testdata/output/
	find . -name .checksum -delete

install-deps:
	go get github.com/spf13/cobra
	go get github.com/stretchr/testify/assert
	go get github.com/mitchellh/go-homedir
	go get github.com/spf13/viper
	go get github.com/aws/aws-sdk-go/aws
	go get github.com/aws/aws-sdk-go/aws/awserr
	go get github.com/aws/aws-sdk-go/aws/session
	go get github.com/aws/aws-sdk-go/service/s3
