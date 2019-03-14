# Multicast

Golang channels are not multicast, but it can be done using a `sync.Cond` & a
`sync.Mutex`. This package defines a `MulticastStream` which captures events
and broadcasts them. It also caches the events so subscribers can replay!
