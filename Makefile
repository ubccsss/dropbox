elections.cgi: main.go
	- go build -v -o dropbox.cgi .

deploy: elections.cgi
	rsync -ravzh --progress . q7w9a@remote.ugrad.cs.ubc.ca:~/public_html/dropbox
