
## 네임스페이스 생성
```
$ kubectl create namespace analytics
```

## Minikube Addon 으로 EFK Enable 설치
minikube 환경에서 편의상 Elasticsearch, Kibana를 설치하는 방법
```
$ minikube addons enable ekf
```

## Helm 으로 Elasticsearch 설치
```
$ helm upgrade --create-namespace --install elasticsearch elastic/elasticsearch -f values.yaml -n analytics
```

## Helm 으로 Kibana 설치
```
$ helm upgrade --create-namespace --install kibana elastic/kibana -n analytics
```