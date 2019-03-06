build:
ifeq ($(version),)
	$(error version is required)
	exit 0
endif
	go build -ldflags "-X main.Version=$(version)" -o bin/fst-core
