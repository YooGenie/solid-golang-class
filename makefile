.PHONEY: install-kafka-docker uninstall-kafka-docker generate-kafka-events
.PHONEY: install-rabbitmq-docker install-rabbitmq-docker generate-rabbitmq-events
.PHONEY: install-elasticsearch-docker uninstall-elasticsearch-docker
.PHONEY: install-kafka-k8s install-prometheus-k8s install-elasticsearch-k8s install-kibana-k8s
.PHONEY: run-local

# docker deployment
## kafka
install-kafka-docker:
	docker compose -f kafka/docker-compose.yaml up -d

uninstall-kafka-docker:
	docker compose -f kafka/docker-compose.yaml down

generate-kafka-events:
	./kafka/gen-events.sh

## rabbitmq
install-rabbitmq-docker:
	docker compose -f rabbitmq/docker-compose.yaml up -d

uninstall-rabbitmq-docker:
	docker compose -f rabbitmq/docker-compose.yaml down

generate-rabbitmq-events:
	cd rabbitmq/mock-data-generator; go run main.go

## elasticsearch

install-elasticsearch-docker:
	docker compose -f elasticsearch/docker-compose.yaml up -d

uninstall-elasticsearch-docker:
	docker compose -f elasticsearch/docker-compose.yaml down

# k8s deployment
install-kafka-k8s:
	kafka/deploy.sh

install-prometheus-k8s:
	prometheus/deploy.sh

install-elasticsearch-k8s:
	cd elasticsearch; ./deploy.sh

install-kibana-k8s:
	kibana/deploy.sh


# build and run
build-local:
	cd app; ./scripts/setup.sh

run-local:
	cd app; go run main.go --debug --config configs/config.yaml