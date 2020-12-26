# go-file
An open-source File Storage System which is developing with golang

# Getting Started

## Dependencies
- You should have Docker installed
- You should place a `.env` file containing base64 encoded public-key

## Start project
Pull the project image:
```shell script
docker pull mjafari98/go-file:latest
```
Run the container:
```shell script
docker run --rm -p 9091:9091 -p 50052:50052 --env-file=.env -it mjafari98/go-file:latest
```