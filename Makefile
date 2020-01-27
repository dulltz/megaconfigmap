IMAGE_NAME = quay.io/dulltz/megaconfigmap-combiner
TAG = `cat TAG`

docker-build:
	docker build -t $(IMAGE_NAME):$(TAG) .

test:
	go test -v -race ./pkg/...
	go vet ./...

e2e:
	go test -v -race ./e2e/...

.PHONY:	docker-build test e2e clean
