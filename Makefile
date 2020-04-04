.PHONY: test docker

DOCKER_IMG = cyrilix/robocar-camera
TAG = latest

test:
	go test ./...

docker:
	docker buildx build . --platform linux/arm/7,linux/arm64,linux/amd64 -t ${DOCKER_IMG}:${TAG} --push

