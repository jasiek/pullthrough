# pullthrough

This is a caching HTTP proxy meant to solver a very particular problem.

Imagine you're running docker, or your deployment process on hundreds of machines requires that you pull the same
exact files from the web over, and over, quite often in parallel. This will cost you bandwidth, time and possibly
money. It's frustrating to have to wait for all of these downloads to complete - that's for sure.

This component acts as a proxy which when it comes across a request will:

* a request for a file in the cache will send that file to the client from disk
* a request for a file not in the cache will start pulling that file from the network and simultaneously stream it to
  the client
* a request for a file that is partially in cache (eg. in progress, see above) will not start any additional
  connections to the outside network, but will share the same connection as initiated via the process above.
