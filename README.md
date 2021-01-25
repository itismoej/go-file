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
docker run --rm -p 50061:50061 --env-file=.env -it mjafari98/go-file:latest
```

## Usage of Examples
### Requirements:
- You should have Go installed on your system

### Upload
You can upload files to the server using command bellow:
```shell
cd examples/upload
go run main.go /path/to/some/file.ext
```

### Download
You can download files from the server using command bellow:
```shell
cd examples/download
go run main.go ID
```
The `ID` should be an integer (1, 2, ...)

The following command will download the file with `ID = 1`, from the server
```shell
go run main.go 1
```
