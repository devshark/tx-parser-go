# Ethereum TX Parser in Go

## Overview of this repository and project

This repository contains the server and client for an ethereum transaction scraper.

The client send their ethereum address to subscribe to, and eventually fetches and stores new transactions since subscription.

## Structure of the project

The project is structured as follows:

- `api` - where project requirements sit and exposed some interfaces for clients to use.
- `client` - where the http client code implements the `api` requirements.
- `app` - where all the app logic are contained, including the server cmd.
- `pkg` - where exported libraries can be shared across the project.

## How to run tests

The tests can be run by executing `make test` from the root directory.

## How to run the server and client

The server codes can be run by executing `make run-server` from the root directory. It will spawn a worker and http server running on port `8080` (or specify by providing `PORT` env).

The client code simply just calls the server http to subscribe desired addresses and then indefinitely fetching the transactions.
