$ CGO_ENABLE=0 GOOS=linux GOARCH=amd64 
SET CGO_ENABLE=0
SET GOOS=linux 
SET GOARCH=amd64

go build -a -ldflags '-extldflags "-static"' .
go build -ldflags='-s -w -extldflags "-static -fpic"'  main.go

go build -ldflags='-s -w -extldflags "-static"' -o main-linux-amd64  ./main.go