# RabbitMQ 설치하고 구동하는 법
---
## Homebrew ( 방법 1)
### 홈브루 업데이트
```
$ brew update
```
### RabbitMQ 설치
```
$ brew install rabbitmq
```

### RabbitMQ 구동
```
$ brew services start rabbitmq
==> Successfully started `rabbitmq` (label: homebrew.mxcl.rabbitmq)
```

### RabbitMQ 서비스 종료
```
$ brew services stop rabbitmq
Stopping `rabbitmq`... (might take a while)
==> Successfully stopped `rabbitmq` (label: homebrew.mxcl.rabbitmq)
```
### RabbitMQ GUI 접속
브라우저를 열고 http://localhost:15672 로 접속합니다.
기본 로그인 `ID/Password`는 `guest/guest` 입니다.

[공식 가이드 문서 원문 링크](https://www.rabbitmq.com/install-homebrew.html)

--- 

## Docker Compose ( 방법 2 )
### RabbitMQ 구동
Docker Desktop 이 기본적으로 설치 되어있어야 합니다.
```
$ docker compose up -d
[+] Running 1/1
 ⠿ Container rabbitmq-rabbitmq-1  Started  
```
### RabbitMQ 서비스 종료
```
$ docker compose up -d
[+] Running 2/2
 ⠿ Container rabbitmq-rabbitmq-1  Removed
 ⠿ Network rabbitmq_default       Removed 
```
### RabbitMQ GUI 접속
브라우저를 열고 http://localhost:15672 로 접속합니다.
기본 로그인 `ID/Password`는 `guest/quest` 입니다.

[공식 가이드 문서 원문 링크](https://hub.docker.com/r/bitnami/rabbitmq)
