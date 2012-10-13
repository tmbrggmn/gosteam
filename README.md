## Intent

gosteam aims to provide a set of functions that interact with various [Steam](http://store.steampowered.com/) interfaces. For example:

* Query the Steam master server to get a list of game servers ([Steam documentation](https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol))
* Query individual Steam/Source game servers for information ([Steam documentation](https://developer.valvesoftware.com/wiki/Server_Queries))
* ...

See the [Features](#features) section for more information.

*Disclaimer*: I'm using this project largely as a playground for learning the [Go](http://golang.org) programming language. It's not at all unlikely that there are things in the codebase that are far from perfect, especially with regards to performance. If you spot a bug or stupidity somewhere, feel free to let me know and I'll try to do better.

## Current state

*   Build status: [![Build Status](https://secure.travis-ci.org/tmbrggmn/gosteam.png)](http://travis-ci.org/tmbrggmn/gosteam)

*   General status: Experimental but working.

    What has been developed has been tested, even though not as thoroughly as I would have liked. *In particular* any scenario that causes a failure hasn't really been tested. For example: what happens when we query a server and the server returns malformed data (probably a null pointer reference somewhere)? What if the server connection times-out in the middle of sending the server list? Stuff like that.

## Usage

Check out the tests or [documentation](#documentation) to see how you can use the functions.

### Documentation

[API documentation](http://go.pkgdoc.org/github.com/tmbrggmn/gosteam) is available on the rather brilliant [pkgdoc.com site](http://go.pkgdoc.org).

## Features

* Query the master server(s) to get a server list (`servers.GetServerList`)
* Query individual servers to get their information (`servers.GetServerInfo`)
* Query individual servers for their player list, including basic player information (`servers.GetPlayerInfo`)

## Packages

*   `servers`

    Contains all functions that relate to the Steam/Source servers.

## Things to do

In order of importance:

1.  ~~Add time-out support to the server query functions so that clients can avoid getting stuck when something goes wrong reading from the UDP connection~~ (added since 29/09/2012)

2.  ~~Add support for A2S_PLAYER server queries~~ (added since 30/09/2012)

3.  Add support for the RCON protocol
    
	It would be particularly cool if we could implement this with a writable and readable channel, where the writeable channel is used to stream commands in real-time. Ordering might be a problem though. Then again, ordering is a problem anyway.

## Known issues

* [GetServerInfo] When querying for server information, there are a number of bytes that contain *Extra Data*. Since it was [not immediately clear from the developer wiki](https://developer.valvesoftware.com/wiki/Talk:Server_Queries#S2A_INFO2_responses_don.27t_match_the_protocol) on how to unpack these extra data bytes, I didn't bother. It seems they contain the server tags (among other information), which would be useful information to unpack, but I haven't been in the mood recently to start to dissect that stuff. The extra data bytes are included in the response struct in their raw form.

