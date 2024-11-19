PORT ?= 8905

build:	
		docker build --build-arg GO_MAP_PORT=$(port) -t go-ip-map-testing .

run:	
		docker-compose up --build -d

stop:
		docker-compose down