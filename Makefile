.PHONY: test bazeltest gotest

BAZEL=bazel
BAZEL_TEST=$(BAZEL) test //:all_tests
GO=go
GO_TEST=$(GO) test -race ./...

test:
	if [ -n "$(shell which $(BAZEL))" ]; \
		then $(BAZEL_TEST); \
		else $(GO_TEST); \
		fi

bazeltest:
	$(BAZEL_TEST)

gotest:
	$(GO_TEST)

deploy:
	$(BAZEL) run //cmd/https-bot:push --config=linux_amd64
