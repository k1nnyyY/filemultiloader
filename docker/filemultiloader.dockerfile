FROM golang:1.21.7
WORKDIR /code
COPY . /code
RUN apt update && apt install -y curl mc
# RUN go build -o main main.go
# CMD ["/code/main"]