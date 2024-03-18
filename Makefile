# Server targets

build-server:
	docker build -t pow_server_image -f Dockerfile_server .

start-server:
	export SERVER_FLAGS=$(FLAGS) && docker-compose up -d server

stop-server:
	docker-compose down server

help-server:
	docker-compose run --rm server ./main --help
	@echo "Example: make start-server FLAGS=\"--difficultyTarget=5000\""

test-server:
	docker-compose run server go test -v ./server/...

# Client targets

build-client:
	docker build -t pow_client_image -f Dockerfile_client .

run-client:
	docker-compose run --rm client ./main $(CLIENT_FLAGS)

help-client:
	docker-compose run --rm client ./main --help
	@echo "Example: make run-client CLIENT_FLAGS=\"--difficultyTarget=5000\""

# Common targets

build: build-server build-client

clean:
	docker-compose down --rmi all