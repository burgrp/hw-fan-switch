.DEFAULT_GOAL := help
.PHONY: requirements deploy

PROJECT_NAME=fan-switch
BUILDER_IMAGE=ghcr.io/burgrp/mpy-builder-esp32c3:latest
PORT=/dev/ttyACM0

CONTAINER_NAME=${PROJECT_NAME}-builder

help:
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

submodule: ## Initializes and updates git submodules
	git submodule init
	git submodule update

start:	## Starts build container
	podman run -it --rm --name ${CONTAINER_NAME} --privileged -v ${PWD}:/project -v /dev:/dev -w /opt/mpy/ports/esp32 -e BOARD=GENERIC_C3_USB -e PORT=${PORT} ${BUILDER_IMAGE}

vanilla: ## Builds and flashes vanilla MicroPython
	podman exec -it ${CONTAINER_NAME} make FROZEN_MANIFEST=/project/manifest-dev.py clean deploy

dev: ## Mount application directory, run main and repl
	podman exec -it -w /project ${CONTAINER_NAME} mpremote reset
	sleep 1
	podman exec -it -w /project ${CONTAINER_NAME} mpremote mount -l app exec "import main" repl

freeze: ## Builds and flashes frozen application
	podman exec -it ${CONTAINER_NAME} make FROZEN_MANIFEST=/project/manifest.py clean deploy

flash: ## Flashes last build (vanilla or freeze)
	podman exec -it ${CONTAINER_NAME} make FROZEN_MANIFEST=/project/manifest.py deploy
