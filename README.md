# robocar-camera

Microservice part to manage camera


## Docker build

```bash
export DOCKER_CLI_EXPERIMENTAL=enabled
docker buildx build . --platform linux/amd64,linux/arm/7,linux/arm64 -t cyrilix/robocar-led
```
