gen-templ:
	@templ generate

run:
	@go run ./cmd/cli.go

runStubServer:
	@go run ./cmd/stubserver.go

runHtmxServer:
	make gen-templ
	@go run ./cmd/htmxserver.go

