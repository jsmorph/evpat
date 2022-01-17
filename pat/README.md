# Event pattern (matching)

A naive implementation of [Amazon EventBridge event
   patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html).
   

## ToDo

- [ ] Null
- [ ] Not
- [ ] CIDR
- [ ] Number comparison is supposed to be by string comparison?

    > For numbers, EventBridge uses string representation. For
    > example, 300, 300.0, and 3.0e2 are not considered equal.
	
- [ ] Many more test cases
- [x] Test cases as JSON
- [ ] Benchmarks
- [x] Array values
- [ ] Better documentation since IMHO the AWS [docs](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html) aren't great
