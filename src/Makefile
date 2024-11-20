APP_NAME=kinozal-bot
SRC=main.go
ZIP=zip
UNAME=$(shell uname)

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux $(SRC)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(APP_NAME).exe $(SRC)

build-macos:
	GOOS=darwin GOARCH=arm64 go build -o $(APP_NAME)-darwin $(SRC)

build-all: build-linux build-windows build-macos

package-linux:
	$(ZIP) $(APP_NAME)-linux.zip $(APP_NAME)-linux .env.example README.md

package-windows:
	$(ZIP) $(APP_NAME)-windows.zip $(APP_NAME).exe .env.example  README.md

package-macos:
	$(ZIP) $(APP_NAME)-macos.zip $(APP_NAME)-darwin .env.example README.md

package-all: package-linux package-windows package-macos

clean:
	rm -f $(APP_NAME)-linux $(APP_NAME)-darwin $(APP_NAME).exe $(APP_NAME)-*.zip