PKGS=./...
COVER_OUT=coverage.out

.PHONY: test clean cover-html

test:
	go test -cover -coverprofile=$(COVER_OUT) $(PKGS)
	@go tool cover -func=$(COVER_OUT) | grep total

cover-html:
	go tool cover -html=$(COVER_OUT) -o coverage.html

clean:
	rm -f $(COVER_OUT) coverage.html
