# castaway

<p align="center">
    <img height="828" src="/.github/castaway.png" alt="screen shot of castaway">
</p>

This is not a peer to peer kind of file transfer solution that utilize
no server in between the clients, and this project is solely for fun
purposes and please... don't use this in production.

It works by streaming byte data from the client to the server and passing it
to the client directly without storing any of it. Because no data is being saved,
we can tell that castaway is pretty lightweight and can handle transferring
very large files efficiently.

## Installations

### Docker (recommended)

```bash
# Build the image
docker build -t castaway .

# Run the container
# you can always remove the rm flag to make the container persistence
docker run --rm -p 6969:6969 --name castaway-test castaway
```

### Manually

Make sure you already installed Node and Golang `1.24` after that you can just do
`make build` to build the binary.

## Acknowledgement

This little project was highly inspired by [Fileway](https://github.com/proofrock/fileway)
