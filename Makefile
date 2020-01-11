IMAGE_NAME = quay.io/dulltz/megaconfigmap-combiner
TAG = `cat TAG`

build:
	docker build -t $(IMAGE_NAME):$(TAG) .

.PHONY:	build
