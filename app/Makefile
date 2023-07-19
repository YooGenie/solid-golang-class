.PHONEY: up-event-data-pipeline-with-kafka, down-event-data-pipeline-with-kafka, gen-event-kafka
.PHONEY: up-event-data-pipeline-with-rabbitmq, down-event-data-pipeline-with-rabbitmq, gen-event-rabbitmq
.PHONEY: helm-install, helm-destroy

up-event-data-pipeline-with-kafka:
	docker compose -f dockercomposes/docker-compose-kafka.yaml -f dockercomposes/docker-compose-kafka-dependencies.yaml up -d
down-event-data-pipeline-with-kafka:
	docker compose -f dockercomposes/docker-compose-kafka.yaml -f dockercomposes/docker-compose-kafka-dependencies.yaml down
gen-kafka-event:
	curl -s "https://api.mockaroo.com/api/d5a195e0?count=1000&key=ff7856d0" | kcat -b localhost:9092 -t purchases -P

up-event-data-pipeline-with-rabbitmq:
	docker compose -f dockercomposes/docker-compose-rabbitmq.yaml -f dockercomposes/docker-compose-rabbitmq-dependencies.yaml up -d
down-event-data-pipeline-with-rabbitmq:
	docker compose -f dockercomposes/docker-compose-rabbitmq.yaml -f dockercomposes/docker-compose-rabbitmq-dependencies.yaml down
gen-rabbitmq-event:
	cd ../rabbitmq/mock-data-generator; go run main.go


kubectl-apply:
	kubectl apply -f manifests/configmap.yaml
	kubectl apply -f manifests/deployment.yaml

helm-install:
	helm upgrade --create-namespace --atomic --install event-data-pipeline helm/ -n events

helm-destroy:
	helm uninstall event-data-pipeline -n events