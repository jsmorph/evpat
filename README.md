# Event pattern matcher

A naive implementation of [Amazon EventBridge event
   patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html).
   
_Warning: This implemention currently departs from EventBridge
patterns in some intentional (and no doubt unintentional) ways._ For
example, a leaf in a pattern here can be a literal expression (not
wrapped in an array).  That particular departure is likely a bad
idea. _This code is subject to sporadic and capricious change._

This repo contains bonus content, which should probably live
elsewhere: A toy message [bus](bus) with a [server-sent events
(SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
[sse](API).  Yes, SSE seems pretty clunky, but it's reasonably
well-supported, simple, and actually supports the basics of what I
happen to want right now.  Crude command-line executable that
subscribes to Redis is [cmd/sser](here).

## References

1. [Amazon EventBridge event
   patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html)
1. [Server-sent events (SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
1. [Sheens pattern matching](https://github.com/Comcast/sheens#pattern-matching)
1. [Peter Norvig's `patmatch`](https://github.com/norvig/paip-lisp/blob/main/lisp/patmatch.lisp)
