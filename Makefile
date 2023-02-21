HOSTNAME=registry.terraform.io
NAMESPACE=novisto
NAME=ecfollowers
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=linux_amd64

default: install

generate:
	go generate ./...

build:
	go build -ldflags -'extldflags "-static"' -tags timetzdata -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

release:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_darwin_arm
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_freebsd_386
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_linux_386
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_linux_amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_linux_arm
	CGO_ENABLED=0 GOOS=openbsd GOARCH=386 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_openbsd_386
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	CGO_ENABLED=0 GOOS=solaris GOARCH=amd64 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_windows_386
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags -'extldflags "-static"' -tags timetzdata -o ./bin/${BINARY}_${VERSION}_windows_amd64

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...
