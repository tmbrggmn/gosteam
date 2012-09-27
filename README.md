## Intent

gosteam aims to provide a set of functions that interact with various [Steam](http://store.steampowered.com/) interfaces. For example:

* Query the Steam master server to get a list of game servers
* Query individual Steam/Source game servers for information
* ...

See the [Features](#features) section for more information.

*Disclaimer*: I'm using this project largely as a playground for learning the [Go]() programming language. It's not at all unlikely that there are things in the codebase that are far from perfect, especially with regards to performance. If you spot a bug or stupidity somewhere, feel free to let me know and I'll try to do better.

## Current state

Experimental but working.

What has been developed has been tested, even though not as thoroughly as I would have liked. *In particular* any scenario that causes a failure hasn't really been tested. For example: what happens when we query a server and the server returns malformed data (probably a null pointer reference somewhere)? What if the server connection times-out in the middle of sending the server list? Stuff like that.

## Usage

Check out the tests or [documentation](#documentation) to see how you can use the functions.

### Documentation

Until I've figured out a way to include the documentation as a stand-alone thing (in a downloadable archive, online or something else) you can get the documentation yourself by downloading the source (using `go get`) and starting the `godoc` server with the following command: `godoc -http :6060`. You can then browse the documentation using your browser on `http://localhost:6060`.

## Features

* Query the master server(s) to get a server list (`servers.GetServerList`)
* Query individual servers to get their information (`servers.GetServerInfo`)

## Packages

*   `servers`

    Contains all functions that relate to the Steam/Source servers.

## Things to do

In order of importance:

1.  Add time-out support to the server query functions so that clients can avoid getting stuck when something goes wrong reading from the UDP connection

2.  Add support for A2S_PLAYER server queries

3.  Add support for the RCON protocol
    
	It would be particularly cool if we could implement this with a writable and readable channel, where the writeable channel is used to stream commands in real-time. Ordering might be a problem though. Then again, ordering is a problem anyway.

