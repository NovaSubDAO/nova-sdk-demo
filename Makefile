public/index.html:
	cd frontend && pnpm i && pnpm build && cp -r out/* ../public/

run: public/index.html
	go run main.go

clean:
	rm -r public/*
