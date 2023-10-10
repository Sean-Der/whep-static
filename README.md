# whep-static

`whep-static` is a simple WHEP server to make testing WHEP clients easier.

### Running

* `git clone https://github.com/sean-der/whep-static.git`
* `cd whep-static`
* `go run main.go`

In the command line you should see

```
Open http://localhost:8080 to access this demo
```

### Generating data

The supplied `Dockerfile` will run ffmpeg and generate sample h264 and ogg files for you.

```bash
docker build --output type=local,dest=. .
```

This produces `output.h264` and `output.ogg`.