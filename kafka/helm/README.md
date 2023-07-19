https://strimzi.io/quickstarts/

# kafka 네임스페이스 생성
```
$ kubectl create namespace kafka
```
## Custom Resource Definition 생성
```
$ kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka
```

## Single Node Kafka & Zookeepr 클러스터 생성
```
$ kubectl apply -f https://strimzi.io/examples/latest/kafka/kafka-persistent-single.yaml -n kafka 
```