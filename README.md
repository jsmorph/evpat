# Event pattern matcher

A naive implementation of [Amazon EventBridge event
   patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html).
   
Also bonus content, which should probably live elsewhere: A toy
message [bus](bus) with a [server-sent events
(SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
[sse](API).  Yes, SSE seems pretty clunky, but it's reasonably
well-supported, simple, and actually supports the basics of what I
happen to want right now.  Command-line executable that subscribes to
Redis is [cmd/sser](here).

## References

1. [Amazon EventBridge event
   patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html)
1. [Server-sent events (SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
1. [Sheens pattern matching](https://github.com/Comcast/sheens#pattern-matching)
1. [Peter Norvig's `patmatch`](https://github.com/norvig/paip-lisp/blob/main/lisp/patmatch.lisp)
