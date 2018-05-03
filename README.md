
This is a simple ipfs dht crawler, written to diagnose
connectivity issues in the network.

To build:
```
make
```

to run:
```
ipfs-crawl
```

This will run indefinitely and generate a file `ipfs-crawl.out` in the
current directory, logging (in json) the results of connection attemps
to peers discovered during the crawl.

If you want detailed dial logs, then invoke the crawler with:
```
ipfs-crawl -log-dial
```
This will log all dial events to `ipfs-crawl-events.json` in the current
directory. But see issue #4 for a wart.

Note: You should make sure your file descriptor ulimit is sufficiently
high to potentially connect to all the reachable peers. The network is
currently small enough for this to be practical.
