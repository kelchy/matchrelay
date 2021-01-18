# matchrelay

## Name

*matchrelay* - match IP addresses and selectively relay them to specific upstream

## Description

Module aims to provide a way to segregate traffic based on source IP of a query similar
to how routers perform source based routing instead of destination domains which coredns
is normally doing.

This module has a dependency on the forward module and support multi proxies and resource
optimizations as with the forward module.

## Syntax

~~~ txt
matchrelay {
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

## Author
Kelvin Chua
kelvin@circles.asia
