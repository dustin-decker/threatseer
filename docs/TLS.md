# Threatseer over mutual TLS

1. Get [certstrap](https://github.com/square/certstrap) to make your certs:
    ```bash
    go get -u github.com/square/certstrap 
    ```
1. Make your CA:
    ```bash
    certstrap init --common-name "threatseer CA"
    ```
1. Make make a request for the `server`:
    ```bash
    certstrap request-cert --domain threatseer
    ```
1. Mint and sign the key and cert for `server`  :
    ```bash
    bin/certstrap sign --CA "threatseer CA" threatseer
    ```
1. Make make a request for the `agent`:
    ```bash
    certstrap request-cert --domain agent
    ```
1. Mint and sign the key and cert for `agent`  :
    ```bash
    bin/certstrap sign --CA "threatseer CA" agent
    ```

The files will be placed in a directory called `out`.