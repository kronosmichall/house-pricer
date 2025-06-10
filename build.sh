cd src
GOOS=linux GOARCH=amd64 go build -o bootstrap .
cd ..
mv src/bootstrap build/
