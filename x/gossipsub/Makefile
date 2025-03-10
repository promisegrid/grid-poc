all: build ship

build:
	go build

ship:
	rsync -Pavx gossipsub $(GOSSIPSUB_REMOTE):/tmp/gossipsub
