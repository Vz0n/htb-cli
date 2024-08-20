all:
	go build -ldflags "-s -w"

clean:
	rm htb-cli
