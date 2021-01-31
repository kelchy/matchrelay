# matchrelay

## Name

*matchrelay* - match IP addresses and selectively relay them to specific upstream

## Description

Module aims to provide a way to segregate traffic based on source IP of a query similar
to how routers perform source based routing instead of destination domains which coredns
is normally doing.

This module has a dependency on the forward module and support multi proxies and resource
optimizations as with the forward module.

to build, pull coredns code
~~~ txt
git clone https://github.com/coredns/coredns.git
~~~

add this line into plugin.cfg

~~~ txt
...
etcd:etcd
loop:loop
matchrelay:github.com/kelchy/matchrelay
forward:forward
grpc:grpc
...
~~~

take note of the order as ordinality of the plugins matter for coredns

since cache is above matchrelay, cache may serve responses without hitting matchrelay
this may cause unexpected behaviours, avoid using cache with matchrelay if the order of
plugins is made this way

you may need to set git to use ssh
~~~ txt
git config --global url."git@github.com:".insteadOf "https://github.com/"
~~~

and set to private
~~~ txt
export GOPRIVATE=github.com/kelchy/matchrelay
~~~

then use "make" to build
~~~ txt
make
~~~

or

~~~ txt
go get github.com/kelchy/matchrelay
go generate
go build
~~~

## Syntax

~~~ txt
matchrelay {
    match ./list.txt
    reload 10s
    net <source ip>
    relay <destination server>
}
~~~

## Examples

Start a server on the default port and load the *matchrelay*

~~~ corefile
example.org {
    matchrelay {
        net 10.1.2.3/32
        relay 8.8.8.8:53 1.1.1.1:53
    }
}
~~~

or by importing a file instead of using the internal
match and reload mechanism. note that if you use reload
module, the whole Corefile will be loaded in each reload.
if the number of zones or list is high, this may cause huge
spikes in CPU which may bring down performance. For very
dynamic environments, use the match and reload mechanism

~~~ corefile
example.org {
    matchrelay {
        import ./list.txt
        relay 8.8.8.8:53 1.1.1.1:53
    }
}
~~~

## Author
Kelvin Chua
kelvin@circles.asia
