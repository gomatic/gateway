.PHONY: test

test: TOKEN := $(shell curl -s localhost:3000/token)
test:
	curl -s -i -H "Authorization: Bearer $(TOKEN)" 'http://localhost:3000/v1/validate'
