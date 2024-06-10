.PHONY: public

public:
	mkdir -p public && cd frontend && pnpm i && pnpm build && cp -r out/* ../public/

run: public
	go run main.go

clean:
	rm -r public/*
