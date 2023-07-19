## Pre-requisites
* Docker Desktop 
https://docs.docker.com/desktop/mac/install/

### Kafka와 Kafka UI 설치
```
$ docker compose up -d
```

## 테스트를 위한 Mock 데이터 생성하기
### kcat 설치
```
$ brew install kcat
```

### kcat 을 이용한 Mockeroo 데이터 생성
```
curl -s "https://api.mockaroo.com/api/d5a195e0?count=1000&key=ff7856d0" | kcat -b localhost:9092 -t purchases -P
```